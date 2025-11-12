package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraDependsOnLastRule(t *testing.T) {
	cases := []struct {
		Name     string
		File     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "resource valid - depends_on last",
			File: "main.tf",
			Content: `
resource "null_resource" "ex" {
  triggers   = { a = 1 }
  depends_on = [null_resource.other]
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "resource invalid - attribute after depends_on",
			File: "main.tf",
			Content: `
resource "null_resource" "ex" {
  depends_on = [null_resource.other]
  triggers   = { a = 1 }
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewStegraDependsOnLastRule(),
					Message: "depends_on must be the last attribute in this block",
					Range: hcl.Range{
						Filename: "main.tf",
						Start:    hcl.Pos{Line: 3, Column: 3},
						End:      hcl.Pos{Line: 3, Column: 37},
					},
				},
			},
		},
		{
			Name: "data invalid - nested block after depends_on not allowed",
			File: "data.tf",
			Content: `
data "aws_iam_policy_document" "p" {
  statement {}
  depends_on = [null_resource.other]
  statement {}
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewStegraDependsOnLastRule(),
					Message: "depends_on must be the last item in this block",
					Range: hcl.Range{
						Filename: "data.tf",
						Start:    hcl.Pos{Line: 4, Column: 3},
						End:      hcl.Pos{Line: 4, Column: 37},
					},
				},
			},
		},
		{
			Name:     "JSON skipped",
			File:     "main.tf.json",
			Content:  `{"resource": {"null_resource": {"x": {"depends_on": []}}}}`,
			Expected: helper.Issues{},
		},
	}

	rule := NewStegraDependsOnLastRule()
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
