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
				rules.NewAwsInstanceExampleTypeRule(),
				rules.NewAwsS3BucketExampleLifecycleRule(),
				rules.NewGoogleComputeSSLPolicyRule(),
				rules.NewTerraformBackendTypeRule(),
				rules.NewTerraformFilenameConventionRule(),
				rules.NewTerraformResourceNameRule(),
			},
		},
	})
}
