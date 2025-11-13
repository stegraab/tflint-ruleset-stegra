package rules

import (
    "path/filepath"
    "strings"

    "github.com/hashicorp/hcl/v2"
    "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraNoLeadingTrailingBlankLinesRule ensures there are no leading or trailing blank lines in files.
type StegraNoLeadingTrailingBlankLinesRule struct{ tflint.DefaultRule }

func NewStegraNoLeadingTrailingBlankLinesRule() *StegraNoLeadingTrailingBlankLinesRule {
    return &StegraNoLeadingTrailingBlankLinesRule{}
}
func (r *StegraNoLeadingTrailingBlankLinesRule) Name() string         { return "stegra_no_leading_trailing_blank_lines" }
func (r *StegraNoLeadingTrailingBlankLinesRule) Enabled() bool        { return true }
func (r *StegraNoLeadingTrailingBlankLinesRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *StegraNoLeadingTrailingBlankLinesRule) Link() string         { return "" }

func (r *StegraNoLeadingTrailingBlankLinesRule) Check(runner tflint.Runner) error {
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
        raw := string(file.Bytes)
        lines := strings.Split(raw, "\n")
        // Precompute line starts (byte offsets)
        lineStarts := make([]int, 1, len(raw)/16+2)
        lineStarts[0] = 0
        for i := 0; i < len(raw); i++ {
            if raw[i] == '\n' {
                lineStarts = append(lineStarts, i+1)
            }
        }

        // Leading blanks: remove every leading blank line
        firstNonBlank := -1
        for i, l := range lines {
            if strings.TrimSpace(strings.TrimRight(l, "\r")) == "" {
                // leading blank, emit fix to delete
                if i+1 < len(lines) {
                    ln := i + 1
                    startByte := lineStarts[i]
                    endByte := lineStarts[i+1]
                    rng := hcl.Range{Filename: filename, Start: hcl.Pos{Line: ln, Column: 1, Byte: startByte}, End: hcl.Pos{Line: ln + 1, Column: 1, Byte: endByte}}
                    if err := runner.EmitIssueWithFix(
                        r,
                        "leading blank lines are not allowed",
                        rng,
                        func(fixer tflint.Fixer) error { return fixer.ReplaceText(rng, "") },
                    ); err != nil {
                        return err
                    }
                }
            } else {
                firstNonBlank = i
                break
            }
        }

        // Trailing blanks: if any exist after the last non-blank line, remove the entire trailing region in one fix
        if firstNonBlank == -1 {
            // File is entirely blank; nothing more to do
            continue
        }
        lastNonBlank := -1
        for i := len(lines) - 1; i >= 0; i-- {
            if strings.TrimSpace(strings.TrimRight(lines[i], "\r")) != "" {
                lastNonBlank = i
                break
            }
        }
        if lastNonBlank >= 0 {
            // trailingCount counts actual blank lines after the last content line, excluding the final empty split element
            trailingCount := (len(lines) - 1) - (lastNonBlank + 1)
            if trailingCount > 0 {
                // Have trailing blank lines starting at lastNonBlank+1 up to EOF
                startLine := lastNonBlank + 2 // 1-based start of first trailing blank
                startIdx := lastNonBlank + 1
                if startIdx < len(lineStarts) {
                    startByte := lineStarts[startIdx]
                    endByte := len(raw)
                    rng := hcl.Range{Filename: filename, Start: hcl.Pos{Line: startLine, Column: 1, Byte: startByte}, End: hcl.Pos{Line: len(lines), Column: 1, Byte: endByte}}
                    if err := runner.EmitIssueWithFix(
                        r,
                        "trailing blank lines are not allowed",
                        rng,
                        // Replace trailing run with exactly one newline to keep a conventional EOF newline
                        func(fixer tflint.Fixer) error { return fixer.ReplaceText(rng, "\n") },
                    ); err != nil {
                        return err
                    }
                }
            }
        }
    }

    return nil
}
