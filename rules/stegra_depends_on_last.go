package rules

import (
    "path/filepath"

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

            msg := "depends_on must be the last item in this block"
            if hasAttrAfter && !hasBlockAfter {
                msg = "depends_on must be the last attribute in this block"
            }
            if err := runner.EmitIssueWithFix(
                r,
                msg,
                depRange,
                func(fixer tflint.Fixer) error {
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
