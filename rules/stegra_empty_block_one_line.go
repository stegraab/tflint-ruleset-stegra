package rules

import (
    "path/filepath"
    "strings"

    "github.com/hashicorp/hcl/v2"
    "github.com/hashicorp/hcl/v2/hclsyntax"
    "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraEmptyBlockOneLineRule enforces that empty blocks use single-line form: `{}`.
// Applies to all block kinds (resource, data, module, and nested blocks).
type StegraEmptyBlockOneLineRule struct{ tflint.DefaultRule }

func NewStegraEmptyBlockOneLineRule() *StegraEmptyBlockOneLineRule { return &StegraEmptyBlockOneLineRule{} }
func (r *StegraEmptyBlockOneLineRule) Name() string                 { return "stegra_empty_block_one_line" }
func (r *StegraEmptyBlockOneLineRule) Enabled() bool                { return true }
func (r *StegraEmptyBlockOneLineRule) Severity() tflint.Severity    { return tflint.ERROR }
func (r *StegraEmptyBlockOneLineRule) Link() string                 { return "" }

func (r *StegraEmptyBlockOneLineRule) Check(runner tflint.Runner) error {
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
        content := string(file.Bytes)

        var walk func(b *hclsyntax.Body) error
        walk = func(b *hclsyntax.Body) error {
            for _, blk := range b.Blocks {
                // Empty if no attributes and no child blocks
                if len(blk.Body.Attributes) == 0 && len(blk.Body.Blocks) == 0 {
                    // Ensure the slice between braces has only whitespace (no comments)
                    start := blk.OpenBraceRange.End.Byte
                    end := blk.CloseBraceRange.Start.Byte
                    if start >= 0 && end >= start && end <= len(content) {
                        between := content[start:end]
                        if strings.TrimSpace(between) == "" {
                            // If braces are not on the same line, fix by removing content between them
                            if blk.CloseBraceRange.Start.Line > blk.OpenBraceRange.End.Line || strings.Contains(between, "\n") {
                                rng := hcl.Range{Filename: filename, Start: hcl.Pos{Byte: start}, End: hcl.Pos{Byte: end}}
                                if err := runner.EmitIssueWithFix(
                                    r,
                                    "empty block must be on one line (use `{}`)",
                                    blk.TypeRange,
                                    func(fixer tflint.Fixer) error { return fixer.ReplaceText(rng, "") },
                                ); err != nil {
                                    return err
                                }
                            }
                        }
                    }
                }
                // Recurse into nested blocks
                if blk.Body != nil {
                    if err := walk(blk.Body); err != nil {
                        return err
                    }
                }
            }
            return nil
        }

        if err := walk(body); err != nil {
            return err
        }
    }
    return nil
}
