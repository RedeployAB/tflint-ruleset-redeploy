package main

import (
	"github.com/RedeployAB/tflint-ruleset-redeploy/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/plugin"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		RuleSet: &tflint.BuiltinRuleSet{
			Name:    "redeploy",
			Version: version,
			Rules: []tflint.Rule{
				rules.NewTerraformBlockFormatRule(),
				rules.NewTerraformFilenameConventionRule(),
				rules.NewTerraformMetaArgumentOrderRule(),
				rules.NewTerraformMetaArgumentFormatRule(),
				rules.NewTerraformResourceNameRule(),
				rules.NewTerraformSourceFormatRule(),
				rules.NewTerraformBasicModuleStructureRule(),
				rules.NewTerraformTagsArgumentRule(),
				rules.NewTerraformProviderSourceOrderRule(),
				rules.NewTerraformLocalsFileRule(),
				rules.NewTerraformConfigBlockFileRule(),
				rules.NewTerraformProviderMinimumMajorVersionRule(),
				rules.NewTerraformResourceArgumentOrderRule(),
				rules.NewTerraformSingleBlankLinesRule(),
				rules.NewTerraformNoLeadingTrailingBlankLinesRule(),
				rules.NewTerraformModuleDependsOnRule(),
				rules.NewTerraformVariableNullableRule(),
				rules.NewTerraformVariableSensitiveRule(),
				rules.NewTerraformOutputSensitiveRule(),
				rules.NewTerraformVariableEphemeralRule(),
				rules.NewTerraformOutputEphemeralRule(),
				rules.NewTerraformVariableArgumentOrderRule(),
				rules.NewTerraformLocalsMirrorAssignmentRule(),
				rules.NewTerraformOutputArgumentOrderRule(),
				rules.NewTerraformOutputFileRule(),
				rules.NewTerraformBlockOrderRule(),
				rules.NewTerraformVariableOrderRule(),
				rules.NewTerraformOutputOrderRule(),
				rules.NewTerraformVariableFileRule(),
			},
		},
	})
}
