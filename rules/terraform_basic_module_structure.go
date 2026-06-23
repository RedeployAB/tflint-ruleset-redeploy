package rules

import (
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

type TerraformBasicModuleStructureRule struct {
	tflint.DefaultRule
}

func NewTerraformBasicModuleStructureRule() *TerraformBasicModuleStructureRule {
	return &TerraformBasicModuleStructureRule{}
}

func (r *TerraformBasicModuleStructureRule) Name() string {
	return "terraform_basic_module_structure"
}

func (r *TerraformBasicModuleStructureRule) Enabled() bool {
	return true
}

func (r *TerraformBasicModuleStructureRule) Severity() tflint.Severity {
	return tflint.WARNING
}

func (r *TerraformBasicModuleStructureRule) Link() string {
	return GetRuleDocLink(r.Name())
}

func (r *TerraformBasicModuleStructureRule) Check(runner tflint.Runner) error {
	requiredFiles := []string{
		FileMain,
		FileVariables,
		FileOutputs,
		FileTerraform,
	}

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	// Find the common directory prefix for all files (module root)
	moduleRoot := findCommonDirectoryPrefix(files)

	// Build a set of base filenames that exist in the module root
	foundFiles := make(map[string]bool)

	for filename := range files {
		// Get the directory of this file
		dir := filepath.Dir(filename)
		base := filepath.Base(filename)

		// Check if this file is in the module root directory
		if dir == moduleRoot {
			foundFiles[base] = true
		}
	}

	for _, required := range requiredFiles {
		if !foundFiles[required] {
			if err := runner.EmitIssue(
				r,
				"Missing required file: "+required,
				hcl.Range{Filename: required},
			); err != nil {
				return err
			}
		}
	}

	return nil
}

// findCommonDirectoryPrefix finds the longest common directory prefix of all file paths
func findCommonDirectoryPrefix(files map[string]*hcl.File) string {
	if len(files) == 0 {
		return "."
	}

	var dirs []string
	for filename := range files {
		dir := filepath.Dir(filename)
		dirs = append(dirs, dir)
	}

	if len(dirs) == 1 {
		return dirs[0]
	}

	// Split first path into components
	firstParts := strings.Split(filepath.Clean(dirs[0]), string(filepath.Separator))

	// Find common prefix with all other paths
	commonParts := firstParts
	for i := 1; i < len(dirs); i++ {
		parts := strings.Split(filepath.Clean(dirs[i]), string(filepath.Separator))
		commonParts = findCommonPrefix(commonParts, parts)
		if len(commonParts) == 0 {
			break
		}
	}

	if len(commonParts) == 0 {
		return "."
	}

	// Rejoin the common parts
	result := strings.Join(commonParts, string(filepath.Separator))
	if result == "" {
		return "."
	}
	return result
}

// findCommonPrefix finds the common prefix between two string slices
func findCommonPrefix(a, b []string) []string {
	common := make([]string, 0, min(len(a), len(b)))
	for i := range common[:cap(common)] {
		if a[i] != b[i] {
			break
		}
		common = append(common, a[i])
	}
	return common
}
