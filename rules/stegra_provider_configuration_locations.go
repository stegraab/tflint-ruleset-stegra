package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// StegraProviderBlockDisallowedDirsRule forbids provider blocks under configured directories.
type StegraProviderConfigurationLocationsRule struct {
	tflint.DefaultRule
}

func NewStegraProviderConfigurationLocationsRule() *StegraProviderConfigurationLocationsRule {
	return &StegraProviderConfigurationLocationsRule{}
}
func (r *StegraProviderConfigurationLocationsRule) Name() string {
	return "stegra_provider_configuration_locations"
}
func (r *StegraProviderConfigurationLocationsRule) Enabled() bool             { return true }
func (r *StegraProviderConfigurationLocationsRule) Severity() tflint.Severity { return tflint.ERROR }
func (r *StegraProviderConfigurationLocationsRule) Link() string              { return "" }

type providerDirsConfig struct {
	Allowed []string `hclext:"allowed_directories,optional"`
}

func (r *StegraProviderConfigurationLocationsRule) Check(runner tflint.Runner) error {
	// Load required config
	cfg := providerDirsConfig{}
	_ = runner.DecodeRuleConfig(r.Name(), &cfg)
	if len(cfg.Allowed) == 0 {
		return fmt.Errorf("%s: allowed_directories option is required; set it in .tflint.hcl rule \"%s\"", r.Name(), r.Name())
	}

	// Normalize configured directories to slash-separated, cleaned prefixes
	allowed := make([]string, 0, len(cfg.Allowed))
	for _, d := range cfg.Allowed {
		if d == "" {
			continue
		}
		d = filepath.ToSlash(filepath.Clean(d))
		allowed = append(allowed, d)
	}

	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		return nil
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, file := range files {
		// Skip JSON
		if strings.HasSuffix(filename, ".tf.json") || filepath.Ext(filename) == ".json" {
			continue
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			continue
		}

		rel := filepath.ToSlash(filepath.Clean(filename))
		// Check if this file is under any allowed directory
		var isAllowed bool
		for _, d := range allowed {
			if isUnderDir(rel, d) {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			for _, blk := range body.Blocks {
				if blk.Type != "provider" {
					continue
				}
				// Highlight the 'provider' keyword
				issueRange := hcl.Range{Filename: filename, Start: blk.TypeRange.Start, End: blk.TypeRange.End}
				msg := fmt.Sprintf("provider block is only allowed under: %s", strings.Join(allowed, ", "))
				if err := runner.EmitIssue(r, msg, issueRange); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isUnderDir(path, dir string) bool {
	if dir == "" {
		return false
	}
	if dir == "." {
		// Treat '.' as repository root files (no directory component)
		return !strings.Contains(path, "/")
	}
	// Ensure dir ends with a slash for boundary match
	d := dir
	if !strings.HasSuffix(d, "/") {
		d += "/"
	}
	// Add leading boundary to avoid matching middle segments incorrectly if dir is relative
	// We operate on cleaned, slash-separated relative paths as provided by TestRunner.
	return path == dir || strings.HasPrefix(path, d)
}
