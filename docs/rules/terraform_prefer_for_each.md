# terraform_prefer_for_each

## What does this rule do?

This rule discourages using the `count` meta-argument to create multiple
near-identical instances, recommending `for_each` instead.

To avoid false positives, it only flags a `count` expression that
**unambiguously** creates more than one instance. This happens in two ways:

- The expression is structurally a multi-instance pattern: a numeric literal
  greater than or equal to `2` (for example `count = 3`) or a `length(...)` call
  (for example `count = length(var.subnets)`), including when either appears in
  a ternary result branch (`count = var.create ? length(var.subnets) : 0`).
- The expression **evaluates** to a known number greater than or equal to `2`
  using Terraform's evaluation context, such as variable defaults, locals, and
  `.tfvars`. For example, `count = var.instance_count` is flagged when
  `instance_count` defaults to `3`.

It deliberately does **not** flag conditional `0`/`1` toggles
(`count = var.create ? 1 : 0`) or a literal `0` or `1`. Expressions whose value
cannot be determined statically, such as a required variable with no default or
a count derived from a resource attribute, are left alone.

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
