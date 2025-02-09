# terraform_single_blank_lines

## What does this rule do?

This rule detects if any file contains more than one consecutive blank line. It ensures that, where a blank line is needed, only a single blank line is used, and no multiple consecutive blank lines appear.

## Why is this important?

Keeping the file free of excessive blank lines improves readability and ensures a consistent, clean style across your Terraform configurations.

## How to fix issues

Remove extra consecutive blank lines so that at most one blank line appears in any location.
