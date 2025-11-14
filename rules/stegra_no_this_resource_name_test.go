package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraNoThisResourceNameRule_FixAndUpdateReferences(t *testing.T) {
	rule := NewStegraNoThisResourceNameRule()
	files := map[string]string{
		"main.tf": `resource "aws_s3_bucket" "this" {}
resource "aws_iam_role" "r" {
  name = aws_s3_bucket.this.id
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
			Message: "resource name must not be 'this' (renamed to 'main' and updated references)",
			Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 1, Column: 26}, End: hcl.Pos{Line: 1, Column: 32}},
		},
	}, runner.Issues)
	helper.AssertChanges(t, map[string]string{
		"main.tf": "resource \"aws_s3_bucket\" \"main\" {}\nresource \"aws_iam_role\" \"r\" {\n  name = aws_s3_bucket.main.id\n}\n",
	}, runner.Changes())
}

func Test_StegraNoThisResourceNameRule_FixWhenNotReferenced(t *testing.T) {
	rule := NewStegraNoThisResourceNameRule()
	files := map[string]string{
		"main.tf": `resource "aws_s3_bucket" "this" {}
`,
	}
	runner := helper.TestRunner(t, files)
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	helper.AssertIssues(t, helper.Issues{
		{
			Rule:    rule,
			Message: "resource name must not be 'this' (renamed to 'main')",
			Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 1, Column: 26}, End: hcl.Pos{Line: 1, Column: 32}},
		},
	}, runner.Issues)
	helper.AssertChanges(t, map[string]string{
		"main.tf": "resource \"aws_s3_bucket\" \"main\" {}\n",
	}, runner.Changes())
}
