# terraform_single_ternary_per_line

## What does this rule do?

This rule checks that a single line contains at most one ternary (conditional)
operation. Any line with more than one ternary triggers the rule, whether the
ternaries are nested or chained (`a ? b : c ? d : e`) or independent
(`[var.a ? 1 : 2, var.b ? 3 : 4]`).

## Why is this important?

Stacking multiple ternary operations on one line makes the logic hard to read
and review. Breaking the logic into intermediate `local` values makes each
decision explicit and self-documenting.

## How to fix issues

Move the nested logic into `local` values so that each line evaluates at most
one ternary.

**Incorrect:**

```hcl
locals {
  nat_gateway_count = var.single_nat_gateway ? 1 : var.one_per_az ? length(var.azs) : local.max_subnet_length
}
```

**Correct:**

```hcl
locals {
  multi_nat_gateway_count = var.one_per_az ? length(var.azs) : local.max_subnet_length
  nat_gateway_count       = var.single_nat_gateway ? 1 : local.multi_nat_gateway_count
}
```

A single ternary on a line is always allowed, including when it is spread across
multiple lines:

```hcl
locals {
  name = (
    var.create
    ? var.name
    : "default"
  )
}
```
