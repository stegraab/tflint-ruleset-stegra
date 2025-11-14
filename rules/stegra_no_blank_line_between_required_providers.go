package rules

import (
    "path/filepath"
    "sort"
    "strings"

    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclsyntax"
    "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraNoBlankLineBetweenRequiredProvidersRule enforces no blank lines between providers inside terraform.required_providers.
type StegraNoBlankLineBetweenRequiredProvidersRule struct{ tflint.DefaultRule }

func NewStegraNoBlankLineBetweenRequiredProvidersRule() *StegraNoBlankLineBetweenRequiredProvidersRule {
    return &StegraNoBlankLineBetweenRequiredProvidersRule{}
}
func (r *StegraNoBlankLineBetweenRequiredProvidersRule) Name() string         { return "stegra_no_blank_line_between_required_providers" }
func (r *StegraNoBlankLineBetweenRequiredProvidersRule) Enabled() bool        { return true }
func (r *StegraNoBlankLineBetweenRequiredProvidersRule) Severity() tflint.Severity {
    return tflint.ERROR
}
func (r *StegraNoBlankLineBetweenRequiredProvidersRule) Link() string { return "" }

func (r *StegraNoBlankLineBetweenRequiredProvidersRule) Check(runner tflint.Runner) error {
    files, err := runner.GetFiles()
    if err != nil {
        return err
    }

    for filename, file := range files {
        if filepath.Ext(filename) != ".tf" {
            continue
        }
        root, ok := file.Body.(*hclsyntax.Body)
        if !ok {
            continue
        }
        content := string(file.Bytes)
        // Build line starts for byte offsets
        lineStarts := make([]int, 1, len(content)/16+2)
        lineStarts[0] = 0
        for i := 0; i < len(content); i++ {
            if content[i] == '\n' {
                lineStarts = append(lineStarts, i+1)
            }
        }

        for _, tf := range root.Blocks {
            if tf.Type != "terraform" {
                continue
            }
            tfBody := tf.Body
            // find required_providers child blocks
            for _, rp := range tfBody.Blocks {
                if rp.Type != "required_providers" {
                    continue
                }
                attrs := rp.Body.Attributes
                if len(attrs) < 2 {
                    continue
                }
                // Sort attributes by start byte for order
                type item struct{
                    name string
                    nameStart hcl.Pos
                    exprEnd hcl.Pos
                }
                items := make([]item, 0, len(attrs))
                for name, a := range attrs {
                    items = append(items, item{name: name, nameStart: a.NameRange.Start, exprEnd: a.Expr.Range().End})
                }
                sort.Slice(items, func(i, j int) bool { return items[i].nameStart.Byte < items[j].nameStart.Byte })

                // For consecutive items, ensure no blank lines exist between prev expr end and next name start
                for i := 1; i < len(items); i++ {
                    prev := items[i-1]
                    next := items[i]
                    startLine := prev.exprEnd.Line + 1
                    endLine := next.nameStart.Line - 1
                    if startLine > endLine {
                        continue
                    }
                    // Collect blank lines in region (ignore comment lines)
                    blanks := []int{}
                    for ln := startLine; ln <= endLine; ln++ {
                        if ln >= 1 && ln <= len(lineStarts) {
                            // Extract line text
                            startByte := lineStarts[ln-1]
                            endByte := len(content)
                            if ln < len(lineStarts) {
                                endByte = lineStarts[ln]
                            }
                            text := strings.TrimRight(content[startByte:endByte], "\n")
                            if strings.TrimSpace(text) == "" { // blank line
                                blanks = append(blanks, ln)
                            }
                        }
                    }
                    if len(blanks) == 0 {
                        continue
                    }
                    // Build fix: remove blank lines, bottom-up
                    if err := runner.EmitIssueWithFix(
                        r,
                        "no blank lines allowed between required_providers entries",
                        hcl.Range{Filename: filename, Start: next.nameStart, End: next.nameStart},
                        func(fixer tflint.Fixer) error {
                            for i := len(blanks) - 1; i >= 0; i-- {
                                ln := blanks[i]
                                start := lineStarts[ln-1]
                                end := 0
                                if ln < len(lineStarts) {
                                    end = lineStarts[ln]
                                } else {
                                    // last line of file; use next name start byte
                                    end = next.nameStart.Byte
                                }
                                rng := hcl.Range{Filename: filename, Start: hcl.Pos{Line: ln, Column: 1, Byte: start}, End: hcl.Pos{Line: ln+1, Column: 1, Byte: end}}
                                if err := fixer.ReplaceText(rng, ""); err != nil {
                                    return err
                                }
                            }
                            return nil
                        },
                    ); err != nil {
                        return err
                    }
                }
            }
        }
    }

    return nil
}

