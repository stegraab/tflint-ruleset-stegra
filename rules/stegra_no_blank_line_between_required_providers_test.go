package rules

import (
    "testing"

    "github.com/hashicorp/hcl/v2"
    "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraNoBlankLineBetweenRequiredProvidersRule(t *testing.T) {
    rule := NewStegraNoBlankLineBetweenRequiredProvidersRule()
    files := map[string]string{
        "main.tf": `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.20.0"
    }
    gitlab = {
      source  = "gitlabhq/gitlab"
      version = "18.5.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}
`,
    }
    runner := helper.TestRunner(t, files)
    if err := rule.Check(runner); err != nil {
        t.Fatalf("Unexpected error occurred: %s", err)
    }
    helper.AssertIssues(t, helper.Issues{
        {
            Rule:    rule,
            Message: "no blank lines allowed between required_providers entries",
            Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 13, Column: 5}, End: hcl.Pos{Line: 13, Column: 5}},
        },
    }, runner.Issues)
}

func Test_StegraNoBlankLineBetweenRequiredProvidersRule_Fix(t *testing.T) {
    rule := NewStegraNoBlankLineBetweenRequiredProvidersRule()
    files := map[string]string{
        "main.tf": `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.20.0"
    }
    gitlab = {
      source  = "gitlabhq/gitlab"
      version = "18.5.0"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}
`,
    }
    runner := helper.TestRunner(t, files)
    if err := rule.Check(runner); err != nil {
        t.Fatalf("Unexpected error occurred: %s", err)
    }
    helper.AssertChanges(t, map[string]string{
        "main.tf": "\nterraform {\n  required_providers {\n    aws = {\n      source  = \"hashicorp/aws\"\n      version = \"6.20.0\"\n    }\n    gitlab = {\n      source  = \"gitlabhq/gitlab\"\n      version = \"18.5.0\"\n    }\n    random = {\n      source  = \"hashicorp/random\"\n      version = \"3.7.2\"\n    }\n  }\n}\n",
    }, runner.Changes())
}
