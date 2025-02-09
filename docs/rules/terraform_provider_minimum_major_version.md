# terraform_provider_minimum_major_version

## What does this rule do?

This rule ensures that when a provider’s version constraint includes a minimum (using `>=` or `>`), an upper bound (using `<` or `<=`) must also be specified. Likewise, if only a maximum is given, a minimum is required. (Constraints that use approximate syntax with `~>`, exact constraints with `=`, or exclusion constraints with `!=` are exempt.)

## Why is this important?

Specifying both minimum and maximum version constraints avoids unexpected provider upgrades and helps maintain compatibility. It ensures that your configuration uses a predictable range of provider versions.

## How to fix issues

If the rule reports an issue, update your provider version constraint to include both a minimum and a maximum. For example, change a constraint like `>= 4.0` to something like `>= 4.0, < 5.0`.
