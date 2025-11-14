package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_StegraBlankLineBetweenBlocksRule(t *testing.T) {
	rule := NewStegraBlankLineBetweenBlocksRule()
	cases := []struct {
		Name     string
		Files    map[string]string
		Expected helper.Issues
	}{
		{
			Name: "adjacent resources - insert blank line",
			Files: map[string]string{
				"main.tf": "resource \"null_resource\" \"a\" {}\nresource \"null_resource\" \"b\" {}\n",
			},
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "blocks must be separated by a blank line",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 9}},
				},
			},
		},
		{
			Name: "resource then module - require blank line",
			Files: map[string]string{
				"main.tf": "resource \"null_resource\" \"a\" {}\nmodule \"m\" { source=\"./m\" }\n",
			},
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "blocks must be separated by a blank line",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 7}},
				},
			},
		},
		{
			Name: "comment right above next same-type block still needs blank before comment",
			Files: map[string]string{
				"main.tf": "resource \"null_resource\" \"a\" {}\n# next resource\nresource \"null_resource\" \"b\" {}\n",
			},
			Expected: helper.Issues{
				{
					Rule:    rule,
					Message: "blocks must be separated by a blank line",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 1}, End: hcl.Pos{Line: 3, Column: 9}},
				},
			},
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

func Test_StegraBlankLineBetweenBlocksRule_Fix(t *testing.T) {
	rule := NewStegraBlankLineBetweenBlocksRule()
	files := map[string]string{
		"main.tf": "resource \"null_resource\" \"a\" {}\nresource \"null_resource\" \"b\" {}\n",
	}
	runner := helper.TestRunner(t, files)
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	helper.AssertChanges(t, map[string]string{
		"main.tf": "resource \"null_resource\" \"a\" {}\n\nresource \"null_resource\" \"b\" {}\n",
	}, runner.Changes())
}

func Test_StegraBlankLineBetweenBlocksRule_Fix_WithCommentBetweenBlocks(t *testing.T) {
	rule := NewStegraBlankLineBetweenBlocksRule()
	files := map[string]string{
		"main.tf": "resource \"null_resource\" \"a\" {}\n# comment about next\nresource \"null_resource\" \"b\" {}\n",
	}
	runner := helper.TestRunner(t, files)
	if err := rule.Check(runner); err != nil {
		t.Fatalf("Unexpected error occurred: %s", err)
	}
	helper.AssertChanges(t, map[string]string{
		"main.tf": "resource \"null_resource\" \"a\" {}\n\n# comment about next\nresource \"null_resource\" \"b\" {}\n",
	}, runner.Changes())
}
