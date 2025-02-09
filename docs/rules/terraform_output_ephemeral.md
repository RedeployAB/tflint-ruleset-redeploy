# terraform_output_ephemeral

## What does this rule do?

This rule checks the usage of the **ephemeral** attribute in output blocks. It
ensures that if the attribute is present, it is not explicitly set to `false`
(since `false` is the default behavior).

## Why is this important?

Specifying `ephemeral = false` is redundant and may lead to confusion. The
correct practice is to omit the attribute entirely if ephemeral behavior is not
desired.

## How to fix issues

Remove the line that sets `ephemeral = false` from your output block. If
ephemeral behavior is needed, set it to `true`; otherwise, simply omit the
attribute.
