package rules

import (
    "testing"

    "github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraProviderConfigurationLocationsRule(t *testing.T) {
    rule := NewStegraProviderConfigurationLocationsRule()

    t.Run("allowed at root", func(t *testing.T) {
        files := map[string]string{
            "main.tf": `provider "aws" {}`,
            ".tflint.hcl": `
rule "stegra_provider_configuration_locations" {
  enabled     = true
  allowed_directories = ["."]
}
`,
        }
        runner := helper.TestRunner(t, files)
        if err := rule.Check(runner); err != nil {
            t.Fatalf("Unexpected error occurred: %s", err)
        }
        helper.AssertIssues(t, helper.Issues{}, runner.Issues)
    })

    t.Run("forbidden under modules", func(t *testing.T) {
        files := map[string]string{
            "modules/net/main.tf": `provider "aws" {}`,
            ".tflint.hcl": `
rule "stegra_provider_configuration_locations" {
  enabled     = true
  allowed_directories = ["env"]
}
`,
        }
        runner := helper.TestRunner(t, files)
        if err := rule.Check(runner); err != nil {
            t.Fatalf("Unexpected error occurred: %s", err)
        }
        // Range positions vary slightly by parser; assert without range for stability
        helper.AssertIssuesWithoutRange(t, helper.Issues{
            {
                Rule:    rule,
                Message: "provider block is only allowed under: env",
            },
        }, runner.Issues)
    })

    t.Run("missing config errors", func(t *testing.T) {
        files := map[string]string{ "modules/x/main.tf": `provider "aws" {}` }
        runner := helper.TestRunner(t, files)
        if err := rule.Check(runner); err == nil {
            t.Fatalf("expected error due to missing directories config, got nil")
        }
    })
}
