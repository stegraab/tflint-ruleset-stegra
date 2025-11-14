package rules

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraNoThisResourceNameRule forbids using "this" as a resource name.
// If safe (no references of <type>.this found), it auto-fixes to "main".
type StegraNoThisResourceNameRule struct{ tflint.DefaultRule }

func NewStegraNoThisResourceNameRule() *StegraNoThisResourceNameRule {
	return &StegraNoThisResourceNameRule{}
}
func (r *StegraNoThisResourceNameRule) Name() string              { return "stegra_no_this_resource_name" }
func (r *StegraNoThisResourceNameRule) Enabled() bool             { return true }
func (r *StegraNoThisResourceNameRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *StegraNoThisResourceNameRule) Link() string              { return "" }

func (r *StegraNoThisResourceNameRule) Check(runner tflint.Runner) error {
	// Only need top-level content to enumerate resources
	body, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{Type: "resource", LabelNames: []string{"type", "name"}, Body: &hclext.BodySchema{}},
		},
	}, &tflint.GetModuleContentOption{ExpandMode: tflint.ExpandModeNone})
	if err != nil {
		return err
	}

	// Preload all .tf file contents and AST bodies for reference scans
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}
	type refHit struct {
		file      string
		startByte int
		endByte   int
	}

	for _, blk := range body.Blocks {
		typ := blk.Labels[0]
		name := blk.Labels[1]
		if strings.ToLower(name) != "this" {
			continue
		}

		// Collect AST-based reference hits for <type>.this traversals in all .tf files of this module
		hits := []refHit{}
		for fname, f := range files {
			if filepath.Ext(fname) != ".tf" {
				continue
			}
			if b, ok := f.Body.(*hclsyntax.Body); ok {
				content := string(f.Bytes)
				// Walk expressions in this file to find traversal exprs of form <type>.this
				hclsyntax.Walk(b, &walkCollector{
					wantType: typ,
					filename: fname,
					content:  content,
					onHit: func(startByte, endByte int) {
						hits = append(hits, refHit{file: fname, startByte: startByte, endByte: endByte})
					},
				})
			}
		}

		// Build range for the name label (second label)
		nameRange := blk.LabelRanges[1]
		// Auto-fix when safe or we can update all references as well
		if err := runner.EmitIssueWithFix(
			r,
			func() string {
				if len(hits) > 0 {
					return "resource name must not be 'this' (renamed to 'main' and updated references)"
				}
				return "resource name must not be 'this' (renamed to 'main')"
			}(),
			hcl.Range{Filename: nameRange.Filename, Start: nameRange.Start, End: nameRange.End},
			func(fixer tflint.Fixer) error {
				// Rename label to "main"
				if err := fixer.ReplaceText(hcl.Range{Filename: nameRange.Filename, Start: nameRange.Start, End: nameRange.End}, "\"main\""); err != nil {
					return err
				}
				for _, h := range hits {
					rng := hcl.Range{Filename: h.file, Start: hcl.Pos{Byte: h.startByte}, End: hcl.Pos{Byte: h.endByte}}
					if err := fixer.ReplaceText(rng, ".main"); err != nil {
						return err
					}
				}
				return nil
			},
		); err != nil {
			return err
		}
	}
	return nil
}

// walkCollector implements hclsyntax.Visitor to collect traversal refs of form <type>.this
type walkCollector struct {
	wantType string
	filename string
	content  string
	onHit    func(startByte int, endByte int)
}

func (w *walkCollector) Enter(node hclsyntax.Node) hcl.Diagnostics {
	var tr hcl.Traversal
	var rng hcl.Range
	switch e := node.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		tr = e.Traversal
		rng = e.Range()
	case *hclsyntax.RelativeTraversalExpr:
		tr = e.Traversal
		rng = e.Range()
	default:
		return nil
	}
	if len(tr) >= 2 {
		if root, ok := tr[0].(hcl.TraverseRoot); ok && root.Name == w.wantType {
			if attr, ok := tr[1].(hcl.TraverseAttr); ok && attr.Name == "this" {
				if rng.Start.Byte >= 0 && rng.End.Byte <= len(w.content) {
					segment := w.content[rng.Start.Byte:rng.End.Byte]
					needle := w.wantType + ".this"
					if idx := strings.Index(segment, needle); idx >= 0 {
						start := rng.Start.Byte + idx + len(w.wantType)
						end := start + len(".this")
						w.onHit(start, end)
					}
				}
			}
		}
	}
	return nil
}

func (w *walkCollector) Exit(node hclsyntax.Node) hcl.Diagnostics { return nil }
