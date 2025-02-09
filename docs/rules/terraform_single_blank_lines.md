# terraform_single_blank_lines

## What does this rule do?

This rule detects if any file contains more than one consecutive empty line. It
ensures that, where a empty line is needed, only a single empty line is used,
and no multiple consecutive empty lines appear.

## Why is this important?

Keeping the file free of excessive empty lines improves readability and ensures
a consistent, clean style across your Terraform configurations.

## How to fix issues

Remove extra consecutive empty lines so that at most one empty line appears in
any location.
