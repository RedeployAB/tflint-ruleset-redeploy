# terraform_output_file

## What does this rule do?

This rule ensures that output blocks are only defined in files that follow the
proper naming convention. Valid file names for outputs are:

- `outputs.tf`
- `outputs.<area>.tf` (for example, `outputs.prod.tf`)

If an output block is found in a file that does not match these patterns, an
error is emitted.

## Why is this important?

Keeping output blocks in dedicated files makes your module more organized and
helps distinguish output definitions from other types of configuration.

## How to fix issues

Move your output blocks into a file that matches one of the valid patterns
(e.g., `outputs.tf` or `outputs.dev.tf`).
