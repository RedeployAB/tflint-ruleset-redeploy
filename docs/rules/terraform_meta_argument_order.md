# terraform_meta_argument_order

## What does this rule do?

This rule validates the order in which meta-arguments appear within resource and
module blocks. It checks three things:

1. **Top meta-arguments** (`provider`, `count`, `for_each`) must appear
  before all regular arguments and blocks.
2. **Bottom meta-arguments** (`lifecycle`, `depends_on`) must appear
  after all regular arguments and blocks.
3. Meta-arguments must follow the expected relative order among themselves.

For **resource** blocks, the expected order is:

1. `provider`
2. `count` or `for_each`
3. _(regular arguments and blocks)_
4. `lifecycle`
5. `depends_on`

For **module** blocks, the expected order is:

1. `count` or `for_each`
2. _(regular arguments and blocks, including `source`)_
3. `depends_on`

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

**Incorrect** (top meta-arg after content):

```hcl
resource "azurerm_role_assignment" "blob_contributor" {
  scope                = each.value.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_databricks_access_connector.this.identity[0].principal_id

  for_each = var.storage_accounts
}
```

**Correct:**

```hcl
resource "azurerm_role_assignment" "blob_contributor" {
  for_each = var.storage_accounts

  scope                = each.value.id
  role_definition_name = "Storage Blob Data Contributor"
  principal_id         = azurerm_databricks_access_connector.this.identity[0].principal_id
}
```

**Incorrect** (bottom meta-arg before content):

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
  count = var.module_count

  source = "./module"

  depends_on = [aws_vpc.example]
}
```

Adjust your configurations to match the expected meta-argument order.
