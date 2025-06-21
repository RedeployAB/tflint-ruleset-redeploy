package main

import (
	"fmt"
	"testing"

	"github.com/RedeployAB/tflint-ruleset-redeploy/rules"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func main() {
	t := &testing.T{}
	
	rule := rules.NewTerraformLocalsMirrorAssignmentRule()
	files := map[string]string{
		"locals.tf": `variable "foo" {}

locals {
	bar = var.foo
}
`,
	}

	runner := helper.TestRunner(t, files)
	
	fmt.Println("Before check:")
	fmt.Printf("Issues: %d\n", len(runner.Issues))
	
	if err := rule.Check(runner); err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	
	fmt.Printf("After check - Issues: %d\n", len(runner.Issues))
	
	changes := runner.Changes()
	fmt.Printf("Changes: %+v\n", changes)
	
	for filename, content := range changes {
		fmt.Printf("File %s:\n%s\n", filename, string(content))
	}
}
