# terraform_module_depends_on

## What does this rule do?

This rule warns when a `module` block contains a `depends_on` meta-argument.
Unlike resources, `module` blocks should not include `depends_on` because
dependency management for modules is handled differently.

## Why is this important?

Using `depends_on` in `module` blocks may lead to unexpected dependency
behavior. Terraform modules inherently understand dependencies based on the
resources and outputs they define. Explicitly adding `depends_on` can complicate
the dependency graph and is generally unnecessary.

## How to fix issues

If the rule reports an issue, remove the `depends_on` meta-argument from the
`module` block.

**Incorrect:**

```hcl
module "example" {
  source     = "./module"
  depends_on = [aws_vpc.example]
}
```

**Correct:**

```hcl
module "example" {
  source = "./module"
  # Remove the depends_on meta-argument
}
```

To manage dependencies, rely on implicit dependencies by passing resource
attributes or IDs to the module inputs.

**Example:**

```hcl
module "example" {
  source   = "./module"
  vpc_id   = aws_vpc.example.id  # This creates an implicit dependency
}
```
