package rules

import (
    "testing"

    "github.com/hashicorp/hcl/v2"
    "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraNoLeadingTrailingBlankLinesRule(t *testing.T) {
    rule := NewStegraNoLeadingTrailingBlankLinesRule()
    cases := []struct {
        Name     string
        File     string
        Content  string
        Expected helper.Issues
    }{
        {
            Name: "leading single blank line is flagged and fixable",
            File: "main.tf",
            Content: "\nresource \"aws_vpc\" \"a\" {}\n",
            Expected: helper.Issues{
                {
                    Rule:    rule,
                    Message: "leading blank lines are not allowed",
                    Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 2, Column: 1}},
                },
            },
        },
        {
            Name: "trailing double blank lines flagged once and fixable",
            File: "main.tf",
            Content: "resource \"aws_vpc\" \"a\" {}\n\n\n",
            Expected: helper.Issues{
                {
                    Rule:    rule,
                    Message: "trailing blank lines are not allowed",
                    Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 4, Column: 1}},
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

