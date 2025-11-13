package rules

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraNoBlockEdgeBlankLinesRule prevents leading and trailing blank lines inside resource/data blocks.
type StegraNoBlockEdgeBlankLinesRule struct{ tflint.DefaultRule }

func NewStegraNoBlockEdgeBlankLinesRule() *StegraNoBlockEdgeBlankLinesRule {
	return &StegraNoBlockEdgeBlankLinesRule{}
}
func (r *StegraNoBlockEdgeBlankLinesRule) Name() string              { return "stegra_no_block_edge_blank_lines" }
func (r *StegraNoBlockEdgeBlankLinesRule) Enabled() bool             { return true }
func (r *StegraNoBlockEdgeBlankLinesRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *StegraNoBlockEdgeBlankLinesRule) Link() string              { return "" }

func (r *StegraNoBlockEdgeBlankLinesRule) Check(runner tflint.Runner) error {
    // Apply in all modules (root and nested)

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
		// build line start byte offsets
		lineStarts := make([]int, 1, len(raw)/16+2)
		lineStarts[0] = 0
		for i := 0; i < len(raw); i++ {
			if raw[i] == '\n' {
				lineStarts = append(lineStarts, i+1)
			}
		}

		walkBodyBlocks(body, func(blk *hclsyntax.Block) {
			// Determine the interior line range of the block
			openLine := blk.OpenBraceRange.End.Line
			closeLine := blk.CloseBraceRange.Start.Line
			startLine := openLine + 1
			endLine := closeLine - 1
			if startLine > endLine {
				return // empty or single-line block
			}

			// Leading edge: blank lines immediately after '{'
			if isBlankLine(lines, startLine) {
				// remove run of blank lines from startLine up to first non-blank or endLine+1
				firstContent := startLine
				for l := startLine; l <= endLine; l++ {
					if strings.TrimSpace(lines[l-1]) != "" {
						break
					}
					firstContent = l + 1
				}
				// Build range [startLine, firstContent)
				if firstContent <= len(lineStarts) {
					startByte := lineStarts[startLine-1]
					endByte := lineStarts[firstContent-1]
					rng := hcl.Range{Filename: filename, Start: hcl.Pos{Line: startLine, Column: 1, Byte: startByte}, End: hcl.Pos{Line: firstContent, Column: 1, Byte: endByte}}
					if err := runner.EmitIssueWithFix(
						r,
						"block must not start with a blank line",
						rng,
						func(fixer tflint.Fixer) error { return fixer.ReplaceText(rng, "") },
					); err != nil {
						return
					}
				}
			}

			// Trailing edge: blank lines immediately before '}'
			if isBlankLine(lines, endLine) {
				// find the last non-blank line moving upwards
				lastContent := endLine
				for l := endLine; l >= startLine; l-- {
					if strings.TrimSpace(lines[l-1]) != "" {
						break
					}
					lastContent = l - 1
				}
				// We want to delete from lastContent+1 to closeLine
				fromLine := lastContent + 1
				if fromLine-1 < len(lineStarts) && closeLine-1 < len(lineStarts) {
					startByte := lineStarts[fromLine-1]
					endByte := lineStarts[closeLine-1]
					rng := hcl.Range{Filename: filename, Start: hcl.Pos{Line: fromLine, Column: 1, Byte: startByte}, End: hcl.Pos{Line: closeLine, Column: 1, Byte: endByte}}
					if err := runner.EmitIssueWithFix(
						r,
						"block must not end with a blank line",
						rng,
						func(fixer tflint.Fixer) error { return fixer.ReplaceText(rng, "") },
					); err != nil {
						return
					}
				}
			}
		})
	}

	return nil
}

func isBlankLine(lines []string, line int) bool {
	if line <= 0 || line > len(lines) {
		return false
	}
	return strings.TrimSpace(lines[line-1]) == ""
}

// walkBodyBlocks recursively visits all blocks within the given body.
func walkBodyBlocks(b *hclsyntax.Body, fn func(*hclsyntax.Block)) {
	for _, blk := range b.Blocks {
		fn(blk)
		// In hclsyntax, blk.Body is already *hclsyntax.Body
		if blk.Body != nil {
			walkBodyBlocks(blk.Body, fn)
		}
	}
}
