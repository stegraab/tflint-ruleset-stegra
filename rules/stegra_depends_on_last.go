package rules

import (
    "path/filepath"
    "strings"

    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclsyntax"
    "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraDependsOnLastRule ensures depends_on is the last attribute in resource/data blocks.
type StegraDependsOnLastRule struct {
	tflint.DefaultRule
}

// NewStegraDependsOnLastRule returns a new rule instance.
func NewStegraDependsOnLastRule() *StegraDependsOnLastRule {
	return &StegraDependsOnLastRule{}
}

// Name returns the rule name.
func (r *StegraDependsOnLastRule) Name() string {
	return "stegra_depends_on_last"
}

// Enabled returns whether the rule is enabled by default.
func (r *StegraDependsOnLastRule) Enabled() bool {
	return true
}

// Severity returns the rule severity.
func (r *StegraDependsOnLastRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link.
func (r *StegraDependsOnLastRule) Link() string {
	return ""
}

// Check validates that depends_on, if present, is the last attribute within resource/data blocks.
func (r *StegraDependsOnLastRule) Check(runner tflint.Runner) error {
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
        lines := strings.Split(raw, "\n")
        // Build line start byte offsets for safe slicing
        lineStarts := make([]int, 1, len(raw)/16+2)
        lineStarts[0] = 0
        for i := 0; i < len(raw); i++ {
            if raw[i] == '\n' {
                lineStarts = append(lineStarts, i+1)
            }
        }

        for _, blk := range body.Blocks {
            if blk.Type != "resource" && blk.Type != "data" && blk.Type != "module" {
                continue
            }

            attrs := blk.Body.Attributes
            dep, ok := attrs["depends_on"]
            if !ok {
                continue
            }

            depRange := hcl.Range{Filename: filename, Start: dep.NameRange.Start, End: dep.Expr.Range().End}
            depEnd := dep.Expr.Range().End

            // Determine if any attribute or child block starts after depends_on ends
            hasAttrAfter := false
            for name, a := range attrs {
                if name == "depends_on" {
                    continue
                }
                start := a.NameRange.Start
                if start.Byte > depEnd.Byte {
                    hasAttrAfter = true
                    break
                }
            }

            hasBlockAfter := false
            for _, child := range blk.Body.Blocks {
                // Compare the start byte of the child block type keyword
                start := child.TypeRange.Start
                if start.Byte > depEnd.Byte {
                    hasBlockAfter = true
                    break
                }
            }

            if !(hasAttrAfter || hasBlockAfter) {
                // Enforce a blank line before the depends_on "section": contiguous comments above depends_on belong to the section
                hasBefore := false
                for name, a := range attrs {
                    if name == "depends_on" {
                        continue
                    }
                    if a.NameRange.Start.Byte < depRange.Start.Byte {
                        hasBefore = true
                        break
                    }
                }
                if !hasBefore {
                    for _, cb := range blk.Body.Blocks {
                        if cb.TypeRange.Start.Byte < depRange.Start.Byte {
                            hasBefore = true
                            break
                        }
                    }
                }
                if hasBefore {
                    // find top of comment section above depends_on
                    topLine := depRange.Start.Line
                    for l := depRange.Start.Line - 1; l >= 1; l-- {
                        s := strings.TrimSpace(lines[l-1])
                        if strings.HasPrefix(s, "#") || strings.HasPrefix(s, "//") {
                            topLine = l
                            continue
                        }
                        break
                    }
                    prevLine := topLine - 1
                    needsBlank := false
                    if prevLine >= 1 && prevLine <= len(lines) {
                        if strings.TrimSpace(lines[prevLine-1]) != "" {
                            needsBlank = true
                        }
                    }
                    if needsBlank {
                        anchor := hcl.Range{Filename: filename, Start: hcl.Pos{Line: topLine, Column: 1, Byte: lineStarts[topLine-1]}, End: hcl.Pos{Line: topLine, Column: 1, Byte: lineStarts[topLine-1]}}
                        if err := runner.EmitIssueWithFix(
                            r,
                            "depends_on must be preceded by a blank line",
                            depRange,
                            func(fixer tflint.Fixer) error { return fixer.InsertTextBefore(anchor, "\n") },
                        ); err != nil {
                            return err
                        }
                    }
                }
                continue
            }

            // Prepare fix: move depends_on to the end of the block (before closing brace)
            // Delete original depends_on including trailing newline
            depStartByte := dep.NameRange.Start.Byte
            endLine := depRange.End.Line
            depEndByte := len(raw)
            if endLine < len(lineStarts) {
                depEndByte = lineStarts[endLine]
            }
            delRange := hcl.Range{Filename: filename, Start: hcl.Pos{Line: depRange.Start.Line, Column: depRange.Start.Column, Byte: depStartByte}, End: hcl.Pos{Line: depRange.End.Line, Column: depRange.End.Column, Byte: depEndByte}}
            moveText := raw[depStartByte:depEndByte]
            insertBefore := hcl.Range{Filename: filename, Start: blk.CloseBraceRange.Start, End: blk.CloseBraceRange.End}
            // Ensure exactly one blank line before the depends_on "section": if comments exist at the end, insert the blank line before the first comment
            closeStartLine := blk.CloseBraceRange.Start.Line
            // find trailing comment group prior to the closing brace
            commentTop := 0
            for l := closeStartLine - 1; l >= 1; l-- {
                s := strings.TrimSpace(lines[l-1])
                if strings.HasPrefix(s, "#") || strings.HasPrefix(s, "//") {
                    commentTop = l
                    continue
                }
                break
            }
            // decide where to ensure blank line: before commentTop if any, else before the inserted depends_on
            ensureLine := closeStartLine
            if commentTop > 0 {
                ensureLine = commentTop
            }
            needInsertBlank := false
            var blankAnchor hcl.Range
            if ensureLine-1 >= 1 && ensureLine-1 <= len(lines) {
                if strings.TrimSpace(lines[ensureLine-2]) != "" {
                    needInsertBlank = true
                    blankAnchor = hcl.Range{Filename: filename, Start: hcl.Pos{Line: ensureLine, Column: 1, Byte: lineStarts[ensureLine-1]}, End: hcl.Pos{Line: ensureLine, Column: 1, Byte: lineStarts[ensureLine-1]}}
                }
            }

            msg := "depends_on must be the last item in this block"
            if hasAttrAfter && !hasBlockAfter {
                msg = "depends_on must be the last attribute in this block"
            }
            if err := runner.EmitIssueWithFix(
                r,
                msg,
                depRange,
                func(fixer tflint.Fixer) error {
                    if needInsertBlank {
                        if err := fixer.InsertTextBefore(blankAnchor, "\n"); err != nil {
                            return err
                        }
                    }
                    if err := fixer.InsertTextBefore(insertBefore, moveText); err != nil {
                        return err
                    }
                    return fixer.ReplaceText(delRange, "")
                },
            ); err != nil {
                return err
            }
        }
    }

    return nil
}
