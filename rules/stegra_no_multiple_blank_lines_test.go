package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraNoMultipleBlankLinesRule(t *testing.T) {
	rule := NewStegraNoMultipleBlankLinesRule()
	cases := []struct {
		Name     string
		File     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "single blank line between blocks allowed",
			File: "main.tf",
			Content: `resource "aws_vpc" "a" {}

resource "aws_vpc" "b" {}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "two blank lines flagged once (second blank)",
			File: "main.tf",
			Content: `resource "aws_vpc" "a" {}


resource "aws_vpc" "b" {}
`,
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "multiple consecutive blank lines are not allowed",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 1}, End: hcl.Pos{Line: 4, Column: 1}},
				},
			},
		},
		{
			Name: "three blank lines flagged twice (second and third)",
			File: "main.tf",
			Content: `resource "aws_vpc" "a" {}



resource "aws_vpc" "b" {}
`,
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "multiple consecutive blank lines are not allowed",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 1}, End: hcl.Pos{Line: 4, Column: 1}},
				},
				{
					Rule:    rule,
					Message: "multiple consecutive blank lines are not allowed",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 1}, End: hcl.Pos{Line: 5, Column: 1}},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{tc.File: tc.Content})
			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}
			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
