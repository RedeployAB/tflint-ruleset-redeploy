# Redeploy TFLint Ruleset

This repository contains a custom ruleset for TFLint.

## Requirements

- TFLint v0.42+
- Go v1.22

## Installation

> [!IMPORTANT] This repository does not contain release binaries yet, so this
> installation will not work. See the "Building the plugin" section to get this
> ruleset working.

You can install the plugin with `tflint --init`. Declare a config in
`.tflint.hcl` as follows:

```hcl
plugin "redeploy" {
  enabled = true

  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-redeploy"
}
```

## Rules

| Name                                 | Description                                                                        | Severity | Enabled | Link |
| ------------------------------------ | ---------------------------------------------------------------------------------- | -------- | ------- | ---- |
| aws_instance_example_type            | Example rule for accessing and evaluating top-level attributes                     | ERROR    | ✔       |      |
| aws_s3_bucket_example_lifecycle_rule | Example rule for accessing top-level/nested blocks and attributes under the blocks | ERROR    | ✔       |      |
| google_compute_ssl_policy            | Example rule with a custom rule config                                             | WARNING  | ✔       |      |
| terraform_backend_type               | Example rule for accessing other than resources                                    | ERROR    | ✔       |      |

## Building the plugin

Clone the repository locally and run the following command:

```shell
make
```

You can easily install the built plugin with the following:

```shell
make install
```

You can run the built plugin like the following:

```shell
$ cat << EOS > .tflint.hcl
plugin "template" {
  enabled = true
}
EOS
$ tflint
```
