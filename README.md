# Redeploy TFLint Ruleset

This repository contains a custom ruleset for TFLint implementing the Redeploy
Terraform
[style guide](https://redeploy.atlassian.net/wiki/spaces/ALZ/pages/508002343/Style+guide).

It is currently a work in progress, rules will be added as they are developed.

## Requirements

- TFLint v0.42+
- Go v1.23

## Installation

You can install the plugin with `tflint --init`. Declare a config in
`.tflint.hcl` as follows:

```hcl
plugin "redeploy" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-redeploy"

  signing_key = <<-KEY
  -----BEGIN PUBLIC KEY-----
  MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEiGFP1JrZTfxZUEGX02dlpNeszF/c
  VEq//nbSPf1EQ+WNWQjPtcY+kaMfiJKLLenh82CaWiKaddr6WPHoGrqYyA==
  -----END PUBLIC KEY-----
  KEY
}
```

> Note: You will need to authenticate with GitHub to download the plugin. You
> can do this by setting the `GITHUB_TOKEN` environment variable to a GitHub
> personal access token with the `read:packages` scope.

## Rules

For a complete list of implemented rules with descriptions and severity levels,
see the [rule documentation](docs/rules/README.md).

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
plugin "redeploy" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-redeploy"
}
EOS
$ tflint
```
