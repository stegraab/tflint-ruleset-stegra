package rules

import (
    "testing"

    "github.com/hashicorp/hcl/v2"
    "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraEmptyBlockOneLineRule(t *testing.T) {
    rule := NewStegraEmptyBlockOneLineRule()
    cases := []struct {
        Name     string
        File     string
        Content  string
        Expected helper.Issues
    }{
        {
            Name: "resource empty two-line -> issue",
            File: "main.tf",
            Content: `
resource "random_uuid" "x" {
}
`,
            Expected: helper.Issues{
                {
                    Rule:    rule,
                    Message: "empty block must be on one line (use `{}`)",
                    Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 9}},
                },
            },
        },
        {
            Name: "resource empty one-line OK",
            File: "ok.tf",
            Content: `resource "random_uuid" "x" {}`,
            Expected: helper.Issues{},
        },
        {
            Name: "nested empty two-line -> issue",
            File: "nested.tf",
            Content: `
resource "null_resource" "ex" {
  triggers = {}

  provisioner "local-exec" {
  }
}
`,
            Expected: helper.Issues{
                {
                    Rule:    rule,
                    Message: "empty block must be on one line (use `{}`)",
                    Range:   hcl.Range{Filename: "nested.tf", Start: hcl.Pos{Line: 5, Column: 3}, End: hcl.Pos{Line: 5, Column: 14}},
                },
            },
        },
        {
            Name: "two-line but with comment -> no issue",
            File: "comment.tf",
            Content: `
resource "null_resource" "ex" {
  # keep comment
}
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

func Test_StegraEmptyBlockOneLineRule_Fix(t *testing.T) {
    rule := NewStegraEmptyBlockOneLineRule()
    files := map[string]string{
        "main.tf": "resource \"random_uuid\" \"x\" {\n}\n",
    }
    runner := helper.TestRunner(t, files)
    if err := rule.Check(runner); err != nil {
        t.Fatalf("Unexpected error occurred: %s", err)
    }
    helper.AssertChanges(t, map[string]string{
        "main.tf": "resource \"random_uuid\" \"x\" {}\n",
    }, runner.Changes())
}
