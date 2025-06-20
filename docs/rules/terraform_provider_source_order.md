# terraform_provider_source_order

## What does this rule do?

This rule validates the ordering of attributes within the `required_providers`
block. It requires that for each provider, the `source` attribute appears before
the `version` attribute.

## Why is this important?

A consistent ordering (with `source` first) improves readability and makes it
easier to review provider configurations. It ensures that version constraints
follow a predictable format.

## How to fix issues

Reorder your provider declaration so that the `source` attribute comes before
the `version` attribute.
