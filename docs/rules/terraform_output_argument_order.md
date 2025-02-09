# terraform_output_argument_order

## What does this rule do?

This rule enforces a specific order for attributes within an output block. The
expected order is:

1. **description**
2. **value**
3. **ephemeral**
4. **sensitive**
5. **precondition**
6. **depends_on**

If any of these attributes appears out of order, an error is emitted.

## Why is this important?

A consistent attribute order in output blocks improves readability and makes it
easier for developers to quickly locate and understand each output’s purpose.

## How to fix issues

Rearrange the attributes in your output block so that they appear in the order
specified above.
