# terraform_resource_argument_order

## What does this rule do?

This rule ensures that within all blocks (resource, data, provider, terraform,
and any nested blocks), all non-block attributes appear before any nested blocks.
(Meta-arguments such as `count`, `for_each`, `depends_on`, `lifecycle`, and
`tags` are ignored by this rule.)

## Why is this important?

Having all regular attributes declared before nested blocks makes configurations
clearer and easier to maintain. It prevents confusion caused by mixing attribute
types.

## How to fix issues

Reorder your blocks so that every non-block attribute is defined before any
nested blocks.

For example, this is incorrect:

```hcl
resource "azurerm_storage_account" "this" {
  name = "storageaccount"

  blob_properties {
    delete_retention_policy {
      days = 7
    }
    versioning_enabled = false  # ERROR: attribute after nested block
  }
}
```

The correct order would be:

```hcl
resource "azurerm_storage_account" "this" {
  name = "storageaccount"

  blob_properties {
    versioning_enabled = false  # Attributes first

    delete_retention_policy {   # Then nested blocks
      days = 7
    }
  }
}
```
