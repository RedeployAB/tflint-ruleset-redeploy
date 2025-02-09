# terraform_tags_argument

## What does this rule do?

This rule validates the placement and spacing of the `tags` argument within resource blocks. It enforces that:

- The `tags` attribute appears after all other non-meta arguments.
- If attributes like `depends_on` or `lifecycle` follow, there must be exactly one blank line separating `tags` from those attributes.

## Why is this important?

Correct placement and spacing of `tags` help maintain a clean and predictable structure in resource blocks. It improves readability and ensures that tag definitions are clearly separated from other configurations.

## How to fix issues

Reorder your resource block so that the `tags` attribute is positioned after all regular arguments and, if necessary, insert exactly one blank line between `tags` and subsequent `depends_on` or `lifecycle` attributes.
