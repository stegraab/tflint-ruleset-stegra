package rules

import (
    "path/filepath"
    "strings"

    "github.com/hashicorp/hcl/v2"
    "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraNoMultipleBlankLinesRule ensures there are no multiple consecutive blank lines anywhere.
type StegraNoMultipleBlankLinesRule struct{ tflint.DefaultRule }

func NewStegraNoMultipleBlankLinesRule() *StegraNoMultipleBlankLinesRule { return &StegraNoMultipleBlankLinesRule{} }
func (r *StegraNoMultipleBlankLinesRule) Name() string                   { return "stegra_no_multiple_blank_lines" }
func (r *StegraNoMultipleBlankLinesRule) Enabled() bool                  { return true }
func (r *StegraNoMultipleBlankLinesRule) Severity() tflint.Severity      { return tflint.ERROR }
func (r *StegraNoMultipleBlankLinesRule) Link() string                   { return "" }

func (r *StegraNoMultipleBlankLinesRule) Check(runner tflint.Runner) error {
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

        // Scan line-by-line and flag any consecutive blank lines beyond the first.
        rawContent := string(file.Bytes)
        lines := strings.Split(rawContent, "\n")
        // Pre-compute byte offsets for the start of each line to build precise ranges
        lineStarts := make([]int, 0, len(lines))
        lineStarts = append(lineStarts, 0)
        for i := 0; i < len(rawContent); i++ {
            if rawContent[i] == '\n' {
                lineStarts = append(lineStarts, i+1)
            }
        }

        blankCount := 0
        seenNonBlank := false
        for i, raw := range lines {
            line := strings.TrimRight(raw, "\r")
            if strings.TrimSpace(line) == "" { // blank line
                // Leading blanks handled by a separate rule; skip until first non-blank
                if !seenNonBlank {
                    continue
                }
                blankCount++
                if blankCount >= 2 {
                    ln := i + 1 // 1-based
                    // Prefer fix: remove one of the redundant blank lines
                    if i+1 < len(lines) {
                        // Remove the current blank line including its newline
                        startByte := lineStarts[i]
                        endByte := lineStarts[i+1]
                        rng := hcl.Range{Filename: filename, Start: hcl.Pos{Line: ln, Column: 1, Byte: startByte}, End: hcl.Pos{Line: ln + 1, Column: 1, Byte: endByte}}
                        if err := runner.EmitIssueWithFix(
                            r,
                            "multiple consecutive blank lines are not allowed",
                            rng,
                            func(fixer tflint.Fixer) error { return fixer.ReplaceText(rng, "") },
                        ); err != nil {
                            return err
                        }
                    } else {
                        // Trailing blanks handled by separate rule; skip
                    }
                }
            } else {
                blankCount = 0
                seenNonBlank = true
            }
        }
    }

    return nil
}
