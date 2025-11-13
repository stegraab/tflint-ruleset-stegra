package rules

import (
    "testing"

    "github.com/hashicorp/hcl/v2"
    "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraNoTypeInNameRule(t *testing.T) {
    rule := NewStegraNoTypeInNameRule()
    cases := []struct {
        Name     string
        File     string
        Content  string
        Expected helper.Issues
    }{
        {
            Name: "resource repeats full type tokens",
            File: "main.tf",
            Content: `
resource "aws_security_group_rule" "my_security_group_rule" {}
`,
            Expected: helper.Issues{
                {
                    Rule:    rule,
                    Message: "resource name `my_security_group_rule` must not repeat type tokens (security, group, rule)",
                    Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 36}, End: hcl.Pos{Line: 2, Column: 60}},
                },
            },
        },
        {
            Name: "resource repeats partial type token",
            File: "main.tf",
            Content: `
resource "aws_security_group_rule" "my_rule" {}
`,
            Expected: helper.Issues{
                {
                    Rule:    rule,
                    Message: "resource name `my_rule` must not repeat type tokens (rule)",
                    Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 36}, End: hcl.Pos{Line: 2, Column: 45}},
                },
            },
        },
        {
            Name: "resource allowed - different name",
            File: "main.tf",
            Content: `
resource "aws_security_group_rule" "app" {}
`,
            Expected: helper.Issues{},
        },
        {
            Name: "data repeats type token",
            File: "data.tf",
            Content: `
data "aws_vpc" "prod_vpc" {}
`,
            Expected: helper.Issues{
                {
                    Rule:    rule,
                    Message: "data name `prod_vpc` must not repeat type tokens (vpc)",
                    Range:   hcl.Range{Filename: "data.tf", Start: hcl.Pos{Line: 2, Column: 16}, End: hcl.Pos{Line: 2, Column: 26}},
                },
            },
        },
        {
            Name: "data allowed - different name",
            File: "data.tf",
            Content: `
data "aws_vpc" "network" {}
`,
            Expected: helper.Issues{},
        },
        {
            Name: "resource allowed - 'main' in type and name",
            File: "main.tf",
            Content: `
resource "aws_main_route_table_association" "main" {}
`,
            Expected: helper.Issues{},
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
