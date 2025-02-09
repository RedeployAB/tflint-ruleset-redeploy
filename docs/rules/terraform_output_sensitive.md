# terraform_output_sensitive

## What does this rule do?

This rule verifies that the **sensitive** attribute in output blocks is not
explicitly set to `false`. Since outputs are not considered sensitive by
default, explicitly setting `sensitive = false` is unnecessary.

## Why is this important?

Omitting the **sensitive** attribute when the output is not sensitive avoids
redundancy and reduces confusion about the sensitivity of the output value.

## How to fix issues

Remove the `sensitive = false` assignment from your output block.
