package rules

import (
    "path/filepath"
    "strings"

    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclsyntax"
    "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraNoBlankLinesInRequiredProvidersRule enforces no blank lines anywhere inside terraform.required_providers.
type StegraNoBlankLinesInRequiredProvidersRule struct{ tflint.DefaultRule }

func NewStegraNoBlankLinesInRequiredProvidersRule() *StegraNoBlankLinesInRequiredProvidersRule {
    return &StegraNoBlankLinesInRequiredProvidersRule{}
}
func (r *StegraNoBlankLinesInRequiredProvidersRule) Name() string {
    return "stegra_no_blank_lines_in_required_providers"
}
func (r *StegraNoBlankLinesInRequiredProvidersRule) Enabled() bool { return true }
func (r *StegraNoBlankLinesInRequiredProvidersRule) Severity() tflint.Severity {
    return tflint.ERROR
}
func (r *StegraNoBlankLinesInRequiredProvidersRule) Link() string { return "" }

func (r *StegraNoBlankLinesInRequiredProvidersRule) Check(runner tflint.Runner) error {
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
                // Scan entire required_providers body for blank lines
                startLine := rp.OpenBraceRange.End.Line + 1
                endLine := rp.CloseBraceRange.Start.Line - 1
                if startLine > endLine {
                    continue
                }
                blanks := []int{}
                for ln := startLine; ln <= endLine; ln++ {
                    if ln >= 1 && ln <= len(lineStarts) {
                        sb := lineStarts[ln-1]
                        eb := len(content)
                        if ln < len(lineStarts) {
                            eb = lineStarts[ln]
                        }
                        text := strings.TrimRight(content[sb:eb], "\n")
                        if strings.TrimSpace(text) == "" {
                            blanks = append(blanks, ln)
                        }
                    }
                }
                if len(blanks) == 0 {
                    continue
                }
                // Remove all blank lines found in this block
                first := blanks[0]
                if err := runner.EmitIssueWithFix(
                    r,
                    "no blank lines allowed in terraform.required_providers",
                    hcl.Range{Filename: filename, Start: hcl.Pos{Line: first, Column: 1}, End: hcl.Pos{Line: first, Column: 1}},
                    func(fixer tflint.Fixer) error {
                        for i := len(blanks) - 1; i >= 0; i-- {
                            ln := blanks[i]
                            start := lineStarts[ln-1]
                            end := 0
                            if ln < len(lineStarts) {
                                end = lineStarts[ln]
                            } else {
                                end = rp.CloseBraceRange.Start.Byte
                            }
                            rng := hcl.Range{Filename: filename, Start: hcl.Pos{Line: ln, Column: 1, Byte: start}, End: hcl.Pos{Line: ln + 1, Column: 1, Byte: end}}
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

    return nil
}
