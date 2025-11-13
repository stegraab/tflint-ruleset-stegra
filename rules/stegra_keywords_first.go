package rules

import (
    "fmt"
    "path/filepath"
    "sort"
    "strings"

    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclsyntax"
    "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraKeywordsFirstRule enforces that configured attributes appear first within resource/data blocks.
type StegraKeywordsFirstRule struct{ tflint.DefaultRule }

func NewStegraKeywordsFirstRule() *StegraKeywordsFirstRule { return &StegraKeywordsFirstRule{} }
func (r *StegraKeywordsFirstRule) Name() string            { return "stegra_keywords_first" }
func (r *StegraKeywordsFirstRule) Enabled() bool           { return true }
func (r *StegraKeywordsFirstRule) Severity() tflint.Severity {
    return tflint.ERROR
}
func (r *StegraKeywordsFirstRule) Link() string { return "" }

type stegraKeywordsFirstConfig struct {
    Keywords []string `hclext:"keywords,optional"`
}

func (r *StegraKeywordsFirstRule) Check(runner tflint.Runner) error {
    // Decode config
    cfg := stegraKeywordsFirstConfig{}
    _ = runner.DecodeRuleConfig(r.Name(), &cfg)
    if len(cfg.Keywords) == 0 {
        return fmt.Errorf("stegra_keywords_first: keywords option is required; set it in .tflint.hcl rule \"stegra_keywords_first\"")
    }
    target := map[string]struct{}{}
    for _, k := range cfg.Keywords {
        target[k] = struct{}{}
    }

    path, err := runner.GetModulePath()
    if err != nil {
        return err
    }
    if !path.IsRoot() {
        return nil
    }

    files, err := runner.GetFiles()
    if err != nil {
        return err
    }

    for filename, file := range files {
        if filepath.Ext(filename) != ".tf" {
            continue
        }
        body, ok := file.Body.(*hclsyntax.Body)
        if !ok {
            continue
        }

        raw := string(file.Bytes)
        // build line start byte offsets for safe slicing
        lineStarts := make([]int, 1, len(raw)/16+2)
        lineStarts[0] = 0
        for i := 0; i < len(raw); i++ {
            if raw[i] == '\n' {
                lineStarts = append(lineStarts, i+1)
            }
        }

        for _, blk := range body.Blocks {
            if blk.Type != "resource" && blk.Type != "data" {
                continue
            }

            // Collect items (attributes and child blocks) in source order
            type item struct {
                kind  string // "attr" or "block"
                name  string // attr name or block type
                start hcl.Pos
                rng   hcl.Range
                startByte int
                endByte   int
            }
            items := make([]item, 0, len(blk.Body.Attributes)+len(blk.Body.Blocks))

            for name, a := range blk.Body.Attributes {
                ir := hcl.Range{Filename: filename, Start: a.NameRange.Start, End: a.Expr.Range().End}
                // Extend end to the start of the next line (so removing includes trailing newline)
                endLine := ir.End.Line
                endByte := len(raw)
                if endLine < len(lineStarts) {
                    endByte = lineStarts[endLine]
                }
                items = append(items, item{kind: "attr", name: name, start: a.NameRange.Start, rng: ir, startByte: ir.Start.Byte, endByte: endByte})
            }
            for _, cb := range blk.Body.Blocks {
                sr := cb.TypeRange
                // End at the start of line after the closing brace
                closeLine := cb.CloseBraceRange.End.Line
                endByte := len(raw)
                if closeLine < len(lineStarts) {
                    endByte = lineStarts[closeLine]
                }
                items = append(items, item{kind: "block", name: cb.Type, start: sr.Start, rng: hcl.Range{Filename: filename, Start: sr.Start, End: sr.End}, startByte: sr.Start.Byte, endByte: endByte})
            }

            sort.Slice(items, func(i, j int) bool { return items[i].start.Byte < items[j].start.Byte })

            // Count how many target attributes are present
            present := 0
            for _, it := range items {
                if it.kind == "attr" {
                    if _, ok := target[it.name]; ok {
                        present++
                    }
                }
            }
            if present == 0 {
                continue
            }

            // Ensure the first 'present' items are all target attributes
            // Pre-compute the index of the first target occurrence for insertion when none seen yet
            firstTargetIdx := -1
            for i, it := range items {
                if it.kind == "attr" {
                    if _, ok := target[it.name]; ok {
                        firstTargetIdx = i
                        break
                    }
                }
            }

            seenTargets := 0
            lastTargetIdx := -1
            for idx, it := range items {
                if it.kind == "attr" {
                    if _, ok := target[it.name]; ok {
                        seenTargets++
                        lastTargetIdx = idx
                        continue
                    }
                }
                if seenTargets < present {
                    // Non-target item appears before all target attributes are listed
                    // Build a fix that moves this item after the last target attribute present
                    off := it // offending item
                    // Compute insertion anchor: before the next item after last target, else before closing brace
                    var insertBefore hcl.Range
                    if lastTargetIdx >= 0 {
                        if lastTargetIdx+1 < len(items) {
                            insertBefore = items[lastTargetIdx+1].rng
                        } else {
                            insertBefore = hcl.Range{Filename: filename, Start: blk.CloseBraceRange.Start, End: blk.CloseBraceRange.End}
                        }
                    } else if firstTargetIdx >= 0 {
                        if firstTargetIdx+1 < len(items) {
                            insertBefore = items[firstTargetIdx+1].rng
                        } else {
                            insertBefore = hcl.Range{Filename: filename, Start: blk.CloseBraceRange.Start, End: blk.CloseBraceRange.End}
                        }
                    } else {
                        // Should not happen because 'present' > 0 guarantees a target; fallback to before '}'
                        insertBefore = hcl.Range{Filename: filename, Start: blk.CloseBraceRange.Start, End: blk.CloseBraceRange.End}
                    }
                    // Define the slice we will move
                    moveText := raw[off.startByte:off.endByte]
                    // Prepare range to delete original
                    delRange := hcl.Range{Filename: filename, Start: hcl.Pos{Line: off.rng.Start.Line, Column: off.rng.Start.Column, Byte: off.startByte}, End: hcl.Pos{Line: off.rng.End.Line, Column: off.rng.End.Column, Byte: off.endByte}}

                    if err := runner.EmitIssueWithFix(
                        r,
                        fmt.Sprintf("These attributes must appear first in this block: %s", strings.Join(cfg.Keywords, ", ")),
                        it.rng,
                        func(fixer tflint.Fixer) error {
                            // Insert moved text before the insertBefore range
                            if err := fixer.InsertTextBefore(insertBefore, moveText); err != nil {
                                return err
                            }
                            // Delete original occurrence
                            return fixer.ReplaceText(delRange, "")
                        },
                    ); err != nil {
                        return err
                    }
                    break
                }
            }
        }
    }

    return nil
}
