# terraform_prefer_for_each

## What does this rule do?

This rule discourages using the `count` meta-argument to create multiple
near-identical instances, recommending `for_each` instead.

To avoid false positives, it only flags a `count` expression that
**unambiguously** creates more than one instance:

- a numeric literal greater than or equal to `2` (for example `count = 3`), or
- a `length(...)` call (for example `count = length(var.subnets)`), including
  when it appears in a ternary result branch
  (`count = var.create ? length(var.subnets) : 0`).

It deliberately does **not** flag conditional `0`/`1` toggles
(`count = var.create ? 1 : 0`), a literal `0` or `1`, or bare references
(`count = var.instance_count`), since those do not necessarily create multiple
instances.

## Why is this important?

Resources created with `count` are tracked by their list index. Inserting or
removing an element shifts every later index, so Terraform plans to destroy and
recreate the trailing resources. `for_each` tracks instances by a stable key,
avoiding this churn. Using `for_each` for collections also produces clearer
plan output.

## How to fix issues

Convert the resource to use `for_each` keyed by a stable identifier.

**Incorrect:**

```hcl
resource "aws_subnet" "this" {
  count = length(var.subnet_cidrs)

  cidr_block = var.subnet_cidrs[count.index]
}
```

**Correct:**

```hcl
resource "aws_subnet" "this" {
  for_each = toset(var.subnet_cidrs)

  cidr_block = each.value
}
```

Using `count` as a conditional toggle (`count = var.create ? 1 : 0`) remains
acceptable and is not reported.
