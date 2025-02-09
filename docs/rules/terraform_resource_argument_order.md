# terraform_resource_argument_order

## What does this rule do?

This rule ensures that within resource (or data, provider, or terraform) blocks, all non-block attributes appear before any nested blocks. (Meta-arguments such as `count`, `for_each`, `depends_on`, `lifecycle`, and `tags` are ignored by this rule.)

## Why is this important?

Having all regular attributes declared before nested blocks makes resource configurations clearer and easier to maintain. It prevents confusion caused by mixing attribute types.

## How to fix issues

Reorder your resource block so that every non-block attribute is defined before any nested blocks.
