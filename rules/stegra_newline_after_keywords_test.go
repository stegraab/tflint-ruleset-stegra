package rules

import (
    "testing"

    "github.com/hashicorp/hcl/v2"
    "github.com/terraform-linters/tflint-plugin-sdk/helper"
    "strings"
)

func Test_StegraNewlineAfterKeywordsRule(t *testing.T) {
	cases := []struct {
		Name     string
		File     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "for_each followed by blank line",
			File: "main.tf",
			Content: `
resource "aws_s3_bucket" "b" {
  for_each = var.buckets

  bucket = each.key
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "for_each missing blank line",
			File: "main.tf",
			Content: `
resource "aws_s3_bucket" "b" {
  for_each = var.buckets
  bucket   = each.key
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewStegraNewlineAfterKeywordsRule(),
					Message: "for_each must be followed by an empty newline",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 25},
					},
				},
			},
		},
		{
			Name: "count followed by blank line (data)",
			File: "data.tf",
			Content: `
data "aws_iam_policy_document" "p" {
  count = 1

  statement { }
}
`,
			Expected: helper.Issues{},
		},
        {
            Name: "module source missing blank line",
			File: "modules.tf",
			Content: `
module "mod" {
  source = "./module"
  version = "~> 1.0"
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewStegraNewlineAfterKeywordsRule(),
					Message: "source must be followed by an empty newline",
					Range: hcl.Range{
						Filename: "modules.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 22},
					},
				},
			},
		},
		{
			Name:     "JSON skipped",
			File:     "main.tf.json",
			Content:  `{"resource": {"null_resource": {"x": {"count": 1}}}}`,
			Expected: helper.Issues{},
        },
        {
            Name: "module source is last (no issue)",
            File: "modules_last.tf",
            Content: `
module "mod" {
  source = "./module"
}
`,
            Expected: helper.Issues{},
        },
        {
            Name: "config overrides keywords (only count)",
            File: "override.tf",
            Content: `
resource "null_resource" "ex" {
  for_each = var.items
  x        = 1
  count    = 1
  y        = 2
}
`,
            Expected: helper.Issues{
                {
                    Rule:    NewStegraNewlineAfterKeywordsRule(),
                    Message: "count must be followed by an empty newline",
                    Range: hcl.Range{
                        Filename: "override.tf",
                        Start:    hcl.Pos{Line: 5, Column: 3},
                        End:      hcl.Pos{Line: 5, Column: 15},
                    },
                },
            },
        },
        {
            Name: "config excludes for_each (no issue)",
            File: "override2.tf",
            Content: `
resource "null_resource" "ex" {
  for_each = var.items
  x        = 1
}
`,
            Expected: helper.Issues{},
        },
}

    rule := NewStegraNewlineAfterKeywordsRule()
    for _, tc := range cases {
        t.Run(tc.Name, func(t *testing.T) {
            files := map[string]string{tc.File: tc.Content}
            if strings.Contains(tc.Name, "config") {
                files[".tflint.hcl"] = `
rule "stegra_newline_after_keywords" {
  enabled  = true
  keywords = ["count"]
}
`
            } else {
                files[".tflint.hcl"] = `
rule "stegra_newline_after_keywords" {
  enabled  = true
  keywords = ["count", "for_each", "source"]
}
`
            }
            runner := helper.TestRunner(t, files)
            if err := rule.Check(runner); err != nil {
                t.Fatalf("Unexpected error occurred: %s", err)
            }
            helper.AssertIssues(t, tc.Expected, runner.Issues)
        })
    }
}

func Test_StegraNewlineAfterKeywordsRule_MissingConfig(t *testing.T) {
    // No .tflint.hcl provided; rule should return an error demanding keywords
    rule := NewStegraNewlineAfterKeywordsRule()
    runner := helper.TestRunner(t, map[string]string{
        "main.tf": `resource "null_resource" "ex" {}`,
    })
    err := rule.Check(runner)
    if err == nil {
        t.Fatalf("expected error due to missing keywords config, got nil")
    }
    if !strings.Contains(err.Error(), "keywords option is required") {
        t.Fatalf("unexpected error: %v", err)
    }
}
