package rules

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraNoTypeInNameRule prevents repeating the type tokens in the name of resources and data sources.
type StegraNoTypeInNameRule struct{ tflint.DefaultRule }

func NewStegraNoTypeInNameRule() *StegraNoTypeInNameRule    { return &StegraNoTypeInNameRule{} }
func (r *StegraNoTypeInNameRule) Name() string              { return "stegra_no_type_in_name" }
func (r *StegraNoTypeInNameRule) Enabled() bool             { return true }
func (r *StegraNoTypeInNameRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *StegraNoTypeInNameRule) Link() string              { return "" }

func (r *StegraNoTypeInNameRule) Check(runner tflint.Runner) error {
	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		return nil
	}

	// Use the module content API to iterate over resource and data blocks.
	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
			{Type: "data", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
		},
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	byType := body.Blocks.ByType()
	for _, kind := range []string{"resource", "data"} {
		for _, block := range byType[kind] {
			typ := strings.ToLower(block.Labels[0])
			name := strings.ToLower(block.Labels[1])

			typeTokens := strings.Split(typ, "_")
			if len(typeTokens) > 1 {
				typeTokens = typeTokens[1:]
			}

			nameTokens := map[string]struct{}{}
			for _, nt := range strings.Split(name, "_") {
				if nt == "" {
					continue
				}
				nameTokens[nt] = struct{}{}
			}

			repeated := make([]string, 0, len(typeTokens))
			for _, tt := range typeTokens {
				if tt == "" {
					continue
				}
				// Allow "main" to be present in both type and name without flagging
				if tt == "main" {
					continue
				}
				if _, ok := nameTokens[tt]; ok {
					repeated = append(repeated, tt)
				}
			}

			if len(repeated) > 0 {
				// Highlight only the name label range (second label)
				issueRange := hcl.Range{Filename: block.LabelRanges[1].Filename, Start: block.LabelRanges[1].Start, End: block.LabelRanges[1].End}
				if err := runner.EmitIssue(
					r,
					kind+" name `"+block.Labels[1]+"` must not repeat type tokens ("+strings.Join(repeated, ", ")+")",
					issueRange,
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
