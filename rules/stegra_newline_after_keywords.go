package rules

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraNewlineAfterKeywordsRule enforces that selected attributes are followed
// by a blank line for readability (e.g., count, for_each, source).
type StegraNewlineAfterKeywordsRule struct {
	tflint.DefaultRule
}

// NewStegraNewlineAfterKeywordsRule returns a new rule
func NewStegraNewlineAfterKeywordsRule() *StegraNewlineAfterKeywordsRule {
	return &StegraNewlineAfterKeywordsRule{}
}

// Name returns the rule name
func (r *StegraNewlineAfterKeywordsRule) Name() string {
	return "stegra_newline_after_keywords"
}

// Enabled returns whether the rule is enabled by default
func (r *StegraNewlineAfterKeywordsRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *StegraNewlineAfterKeywordsRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *StegraNewlineAfterKeywordsRule) Link() string {
	return ""
}

// stegraNewlineConfig allows configuring which keywords to enforce.
type stegraNewlineConfig struct {
	Keywords []string `hclext:"keywords,optional"`
}

// Check scans HCL files and enforces a blank line after target attributes.
func (r *StegraNewlineAfterKeywordsRule) Check(runner tflint.Runner) error {
	// Load configurable keywords; fail if not provided.
	cfg := stegraNewlineConfig{}
	_ = runner.DecodeRuleConfig(r.Name(), &cfg)
	keys := cfg.Keywords
	if len(keys) == 0 {
		return fmt.Errorf("stegra_newline_after_keywords: keywords option is required; set it in .tflint.hcl rule \"stegra_newline_after_keywords\"")
	}
	target := make(map[string]struct{}, len(keys))
	for _, k := range keys {
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
		if filepath.Ext(filename) == ".json" {
			continue
		}
		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		content := file.Bytes
		lines := strings.Split(string(content), "\n")
		newlineSeq := detectLineEnding(content)

		for _, blk := range body.Blocks {
			for name, attr := range blk.Body.Attributes {
				if _, want := target[name]; !want {
					continue
				}
				nameStart := attr.NameRange.Start
				exprRange := attr.Expr.Range()
				ar := hcl.Range{Filename: filename, Start: nameStart, End: exprRange.End}
				endLine := ar.End.Line
				nextIdx := endLine
				notBlank := false
				if nextIdx >= len(lines) {
					notBlank = true
				} else if strings.TrimSpace(strings.TrimRight(lines[nextIdx], "\r")) != "" {
					notBlank = true
				}
				if notBlank {
					if err := runner.EmitIssueWithFix(
						r,
						fmt.Sprintf("%s must be followed by an empty newline", name),
						ar,
						func(fixer tflint.Fixer) error { return fixer.InsertTextAfter(ar, newlineSeq) },
					); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func detectLineEnding(content []byte) string {
	if bytes.Contains(content, []byte("\r\n")) {
		return "\r\n"
	}
	return "\n"
}
