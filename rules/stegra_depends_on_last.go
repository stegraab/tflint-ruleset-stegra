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

        for _, blk := range body.Blocks {
            if blk.Type != "resource" && blk.Type != "data" {
                continue
            }

            attrs := blk.Body.Attributes
            dep, ok := attrs["depends_on"]
            if !ok {
                continue
            }

            depRange := hcl.Range{Filename: filename, Start: dep.NameRange.Start, End: dep.Expr.Range().End}
            depEnd := dep.Expr.Range().End

            // Ensure no other attribute starts after depends_on ends
            for name, a := range attrs {
                if name == "depends_on" {
                    continue
                }
                start := a.NameRange.Start
                if start.Byte > depEnd.Byte {
                    if err := runner.EmitIssue(
                        r,
                        "depends_on must be the last attribute in this block",
                        depRange,
                    ); err != nil {
                        return err
                    }
                    break
                }
            }

            // Ensure no nested block starts after depends_on ends
            for _, child := range blk.Body.Blocks {
                // Compare the start byte of the child block type keyword
                start := child.TypeRange.Start
                if start.Byte > depEnd.Byte {
                    if err := runner.EmitIssue(
                        r,
                        "depends_on must be the last item in this block",
                        depRange,
                    ); err != nil {
                        return err
                    }
                    break
                }
            }
        }
    }

    return nil
}
