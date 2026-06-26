# Contributing

Thanks for your interest in contributing to the Redeploy TFLint ruleset.

## Prerequisites

- Go (see the version in `go.mod`)
- [TFLint](https://github.com/terraform-linters/tflint) for end-to-end tests

## Development

The plugin is written in Go and built with `make`. Common targets:

- `make build` builds the plugin binary
- `make install` builds and installs it into `~/.tflint.d/plugins`
- `make test` runs the unit tests
- `make e2e` builds, installs, and runs the integration tests
- `make lint` runs `golangci-lint`
- `make fmt` formats the code

Before opening a pull request, make sure `make test`, `make lint`, and
`make e2e` all pass.

## Adding a rule

Each rule lives in its own file under `rules/`. To add one:

1. Implement the rule in `rules/terraform_<name>.go`.
2. Register it in `main.go`.
3. Add a table-driven test and fixtures under `rules/testdata/`.
4. Document it in `docs/rules/` and add a row to `docs/rules/README.md`.

Rules should be deterministic and avoid false positives. Prefer reporting only
what is unambiguous over catching every possible case. This ruleset does not
overlap with the official TFLint rulesets, so please do not duplicate rules they
already provide.

## Commit messages and pull requests

This project follows [Conventional Commits][conventional-commits]. Commit
messages and pull request titles are linted, so use a type such as `feat`,
`fix`, `refactor`, `docs`, `test`, or `chore`, keep the subject in lowercase,
and mark breaking changes with `!` (for example `feat!:`).

[conventional-commits]: https://www.conventionalcommits.org/
