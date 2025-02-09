# terraform_meta_argument_order

## What does this rule do?

This rule validates the order in which meta-arguments appear within resource and
module blocks.

For **resource** blocks, the expected order is:

1. `provider`
2. `count` or `for_each`
3. `lifecycle`
4. `depends_on`

For **module** blocks, the expected order is:

1. `count` or `for_each`
2. `depends_on`

If meta-arguments appear out of this order, the rule emits an error.

## Why is this important?

A consistent ordering of meta-arguments makes your Terraform code easier to read
and maintain. It ensures that configurations are predictable and that
meta-arguments are consistently positioned, which aids in understanding the
behavior and dependencies of resources and modules.

## How to fix issues

Reorder the meta-arguments in your resource or module block to follow the
expected sequence.

**For resource blocks:**

**Incorrect:**

```hcl
resource "aws_instance" "example" {
  depends_on    = [aws_vpc.example]
  count         = var.instance_count
  instance_type = "t2.micro"
  ami           = "ami-123456"
}
```

**Correct:**

```hcl
resource "aws_instance" "example" {
  count = var.instance_count

  ami           = "ami-123456"
  instance_type = "t2.micro"

  depends_on = [aws_vpc.example]
}
```

**For module blocks:**

**Incorrect:**

```hcl
module "example" {
  source     = "./module"
  depends_on = [aws_vpc.example]
  count      = var.module_count
}
```

**Correct:**

```hcl
module "example" {
  source = "./module"
  count  = var.module_count

  depends_on = [aws_vpc.example]
}
```

Adjust your configurations to match the expected meta-argument order.
