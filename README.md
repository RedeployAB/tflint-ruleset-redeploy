# Redeploy TFLint Ruleset

This repository contains a custom ruleset for TFLint used for enforcing the
Redeploy Terraform [style guide](https://redeploy.atlassian.net/wiki/spaces/ALZ/pages/508002343/Style+guide).

It is currently a work in progress, rules will be added as they are developed.

## Requirements

- TFLint v0.42+
- Go v1.23

## Installation

> [!IMPORTANT]
> This repository does not contain release binaries yet, so this
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

This section will be added soon. Until then, refer to [`main.go`](main.go) to
see a list of rules that are currently implemented.

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
