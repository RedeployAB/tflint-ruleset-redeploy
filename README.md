# Redeploy TFLint Ruleset

This repository contains a custom ruleset for TFLint implementing the Redeploy
Terraform
[style guide](https://redeploy.atlassian.net/wiki/spaces/ALZ/pages/508002343/Style+guide).

It is currently a work in progress, rules will be added as they are developed.

## Requirements

- TFLint v0.46+
- Go v1.24

## Installation

You can install the plugin with `tflint --init`. Declare a config in
`.tflint.hcl` as follows:

```hcl
plugin "redeploy" {
  enabled = true
  version = "0.2.0"
  source  = "github.com/RedeployAB/tflint-ruleset-redeploy"

  signing_key = <<-KEY
  -----BEGIN PGP PUBLIC KEY BLOCK-----

  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
  gjYRzmMmFDXFqg==
  =q4Rt
  -----END PGP PUBLIC KEY BLOCK-----
  KEY
}
```

> Note: You will need to authenticate with GitHub to download the plugin. You
> can do this by setting the `GITHUB_TOKEN` environment variable to a GitHub
> personal access token with the `read:packages` scope.

> Breaking change: Support for TFLint v0.45 and earlier has been dropped. You must use TFLint v0.46 or newer.

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
  version = "0.2.0"
  source  = "github.com/RedeployAB/tflint-ruleset-redeploy"
}
EOS
$ tflint
```
