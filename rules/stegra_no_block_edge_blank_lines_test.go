package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraNoBlockEdgeBlankLinesRule(t *testing.T) {
	rule := NewStegraNoBlockEdgeBlankLinesRule()
	cases := []struct {
		Name     string
		File     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name:    "resource with leading blank line inside",
			File:    "main.tf",
			Content: "resource \"aws_vpc\" \"a\" {\n\nname = \"a\"\n}\n",
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "block must not start with a blank line",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 3, Column: 1}},
				},
			},
		},
		{
			Name:    "data with trailing blank line inside",
			File:    "data.tf",
			Content: "data \"aws_vpc\" \"a\" {\nname = \"a\"\n\n}\n",
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "block must not end with a blank line",
					Range:   hcl.Range{Filename: "data.tf", Start: hcl.Pos{Line: 3, Column: 1}, End: hcl.Pos{Line: 4, Column: 1}},
				},
			},
		},
		{
			Name:    "both leading and trailing blank lines flagged",
			File:    "main.tf",
			Content: "resource \"aws_vpc\" \"a\" {\n\nname = \"a\"\n\n}\n",
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "block must not start with a blank line",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 3, Column: 1}},
				},
				{
					Rule:    rule,
					Message: "block must not end with a blank line",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 1}, End: hcl.Pos{Line: 5, Column: 1}},
				},
			},
		},
		{
			Name:     "no leading or trailing blank lines inside",
			File:     "ok.tf",
			Content:  "resource \"aws_vpc\" \"a\" {\nname = \"a\"\n}\n",
			Expected: helper.Issues{},
		},
		{
			Name:    "module block leading blank line",
			File:    "mod.tf",
			Content: "module \"vpc\" {\n\nsource = \"./vpc\"\n}\n",
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "block must not start with a blank line",
					Range:   hcl.Range{Filename: "mod.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 3, Column: 1}},
				},
			},
		},
		{
			Name:    "provider block trailing blank line",
			File:    "prov.tf",
			Content: "provider \"aws\" {\nregion = \"eu-west-1\"\n\n}\n",
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "block must not end with a blank line",
					Range:   hcl.Range{Filename: "prov.tf", Start: hcl.Pos{Line: 3, Column: 1}, End: hcl.Pos{Line: 4, Column: 1}},
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
