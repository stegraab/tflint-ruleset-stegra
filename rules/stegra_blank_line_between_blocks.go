package rules

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraBlankLineBetweenBlocksRule enforces at least one blank line between resource, data, and module blocks.
type StegraBlankLineBetweenBlocksRule struct{ tflint.DefaultRule }

func NewStegraBlankLineBetweenBlocksRule() *StegraBlankLineBetweenBlocksRule {
	return &StegraBlankLineBetweenBlocksRule{}
}
func (r *StegraBlankLineBetweenBlocksRule) Name() string              { return "stegra_blank_line_between_blocks" }
func (r *StegraBlankLineBetweenBlocksRule) Enabled() bool             { return true }
func (r *StegraBlankLineBetweenBlocksRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *StegraBlankLineBetweenBlocksRule) Link() string              { return "" }

func (r *StegraBlankLineBetweenBlocksRule) Check(runner tflint.Runner) error {
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
		// Build line start byte offsets so we can construct precise anchors
		lineStarts := make([]int, 1, len(raw)/16+2)
		lineStarts[0] = 0
		for i := 0; i < len(raw); i++ {
			if raw[i] == '\n' {
				lineStarts = append(lineStarts, i+1)
			}
		}

		// Filter to only the block types we care about and keep source order
		blocks := make([]*hclsyntax.Block, 0, len(body.Blocks))
		for _, b := range body.Blocks {
			switch b.Type {
			case "resource", "data", "module":
				blocks = append(blocks, b)
			default:
				// ignore other block types
			}
		}

		for i := 1; i < len(blocks); i++ {
			prev := blocks[i-1]
			next := blocks[i]
			prevCloseLine := prev.CloseBraceRange.End.Line
			nextStartLine := next.TypeRange.Start.Line
			// Analyze the region between the two blocks
			regionStart := prevCloseLine + 1
			regionEnd := nextStartLine - 1
			hasBlank := false
			firstNonBlank := 0
			for ln := regionStart; ln <= regionEnd; ln++ {
				if ln >= 1 && ln <= len(lines) {
					if strings.TrimSpace(lines[ln-1]) == "" {
						hasBlank = true
						break
					}
					if firstNonBlank == 0 && strings.TrimSpace(lines[ln-1]) != "" {
						firstNonBlank = ln
					}
				}
			}
			if !hasBlank {
				// If there is a comment group right before the next block, we still require a blank line,
				// but insert it before the first comment so comments stay attached to the next block.
				// Decide insertion anchor: if we have any line between blocks, insert before the first non-blank line,
				// otherwise insert before the next block itself.
				anchor := next.TypeRange
				if firstNonBlank != 0 {
					// Build a zero-width anchor at the start of the firstNonBlank line with correct byte offsets
					byteOff := 0
					if firstNonBlank-1 < len(lineStarts) {
						byteOff = lineStarts[firstNonBlank-1]
					}
					anchor = hcl.Range{Filename: filename, Start: hcl.Pos{Line: firstNonBlank, Column: 1, Byte: byteOff}, End: hcl.Pos{Line: firstNonBlank, Column: 1, Byte: byteOff}}
				}
				if err := runner.EmitIssueWithFix(
					r,
					"blocks must be separated by a blank line",
					next.TypeRange,
					func(fixer tflint.Fixer) error { return fixer.InsertTextBefore(anchor, "\n") },
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
