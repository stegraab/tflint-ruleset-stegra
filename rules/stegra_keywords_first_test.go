package rules

import (
    "testing"

    "github.com/hashicorp/hcl/v2"
    "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraKeywordsFirstRule(t *testing.T) {
    rule := NewStegraKeywordsFirstRule()
    cfg := `
rule "stegra_keywords_first" {
  enabled  = true
  keywords = ["for_each", "count"]
}
`
    cases := []struct {
        Name     string
        Files    map[string]string
        Expected helper.Issues
    }{
        {
            Name: "for_each first is allowed",
            Files: map[string]string{
                ".tflint.hcl": cfg,
                "main.tf":     "resource \"aws_vpc\" \"a\" {\nfor_each = []\nname = \"a\"\n}\n",
            },
            Expected: helper.Issues{},
        },
        {
            Name: "non-target before target is flagged",
            Files: map[string]string{
                ".tflint.hcl": cfg,
                "main.tf":     "resource \"aws_vpc\" \"a\" {\nname = \"a\"\nfor_each = []\n}\n",
            },
            Expected: helper.Issues{
                {
                    Rule:    rule,
                    Message: "These attributes must appear first in this block: for_each, count",
                    Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 11}},
                },
            },
        },
        {
            Name: "multiple targets can be first in any order",
            Files: map[string]string{
                ".tflint.hcl": cfg,
                "main.tf":     "resource \"aws_vpc\" \"a\" {\ncount = 1\nfor_each = []\nname = \"a\"\n}\n",
            },
            Expected: helper.Issues{},
        },
        {
            Name: "no targets present - no issue",
            Files: map[string]string{
                ".tflint.hcl": cfg,
                "data.tf":     "data \"aws_vpc\" \"a\" {\nname = \"a\"\n}\n",
            },
            Expected: helper.Issues{},
        },
    }

    for _, tc := range cases {
        t.Run(tc.Name, func(t *testing.T) {
            runner := helper.TestRunner(t, tc.Files)
            if err := rule.Check(runner); err != nil {
                t.Fatalf("Unexpected error occurred: %s", err)
            }
            helper.AssertIssues(t, tc.Expected, runner.Issues)
        })
    }
}

func Test_StegraKeywordsFirstRule_Fix_ReordersAttributes(t *testing.T) {
    rule := NewStegraKeywordsFirstRule()
    cfg := `
rule "stegra_keywords_first" {
  enabled  = true
  keywords = ["for_each", "count"]
}
`
    files := map[string]string{
        ".tflint.hcl": cfg,
        "main.tf":     "resource \"aws_vpc\" \"a\" {\nname = \"a\"\nfor_each = []\n}\n",
    }
    runner := helper.TestRunner(t, files)
    if err := rule.Check(runner); err != nil {
        t.Fatalf("Unexpected error occurred: %s", err)
    }
    helper.AssertChanges(t, map[string]string{
        "main.tf": "resource \"aws_vpc\" \"a\" {\n  for_each = []\n  name     = \"a\"\n}\n",
    }, runner.Changes())
}
