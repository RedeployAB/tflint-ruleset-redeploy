# terraform_source_format

## What does this rule do?

This rule enforces proper formatting within module blocks, specifically for the
`source` (and optionally `version`) attributes. It ensures that there is proper
spacing (no unexpected empty lines) after these attributes before the block
ends.

## Why is this important?

A consistent layout in module source declarations helps maintain clarity. It
prevents misinterpretation of version constraints and supports a uniform style.

## How to fix issues

Adjust your module block so that there are no unexpected empty lines after the
`source` or `version` lines. Ensure the attributes and any subsequent properties
are properly spaced.
