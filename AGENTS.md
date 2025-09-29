# Repository Agent Guide

Welcome to the tflint-ruleset-redeploy repository. This guide summarizes best
practices for agents working on this codebase. For repository structure,
requirements, and installation details, see the [README.md](README.md).

## Development Setup

This is a Go project implementing a custom TFLint ruleset for the Redeploy
Terraform style guide. Ensure you have the following prerequisites:

- Go v1.24+
- TFLint v0.46+
- golangci-lint v2.1.5+

## Building the Project

Build the plugin binary:

```bash
make build
```

Install the built plugin locally:

```bash
make install
```

## Commit Messages

Use the
[Conventional Commits specification](https://www.conventionalcommits.org/en/v1.0.0/)
for all commit messages. The repository uses commitlint to enforce this.
Examples:

```text
feat: add terraform_resource_name rule
fix: correct validation logic for output arguments
docs: update rule documentation for terraform_locals_file
test: add test cases for meta argument ordering
chore: update golangci-lint configuration
```

## Pull Request Titles

Use the same Conventional Commits format for PR titles.

## Testing

**Always run tests before finishing any changes.** The project has comprehensive
unit and integration tests:

### Unit Tests

Run unit tests:

```bash
make test
```

### End-to-End Tests

The end-to-end tests require the plugin to be installed. The `e2e` Makefile
target runs `make install` automatically, so running `make e2e` will build and
install the plugin before executing the tests.

```bash
make e2e
```

## Linting

**Always run golangci-lint and fix any issues before finishing:**

```bash
golangci-lint run
```

The project has a `.golangci.yaml` configuration file that defines the linting
rules. Ensure your code passes all linting checks.

## Pre-commit Checklist

Before marking any task as complete, ensure:

1. ✅ All code changes follow Go best practices
2. ✅ `golangci-lint run` passes with no errors
3. ✅ `make test` passes all unit tests
4. ✅ `make e2e` passes all integration tests (this command automatically
    builds and installs the plugin)
5. ✅ Documentation is updated if adding/modifying rules
6. ✅ Commit messages follow Conventional Commits format

## Working with Rules

When adding or modifying TFLint rules:

1. Rule implementations go in the `rules/` directory
2. Each rule should have a corresponding test file (e.g.,
  `terraform_rule_name.go` and `terraform_rule_name_test.go`)
3. Test data files go in `rules/testdata/`
4. Update the rule documentation in `docs/rules/` when adding new rules
5. Follow the existing patterns for rule implementation and testing

## Common Commands Summary

```bash
# Build the plugin
make build

# Install locally
make install

# Run unit tests
make test

# Run integration tests (plugin will be built and installed automatically)
make e2e

# Run linter
golangci-lint run

# Run all checks (recommended before pushing)
golangci-lint run && make test && make e2e
```
