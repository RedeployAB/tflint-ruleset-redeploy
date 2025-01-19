provider "azurerm" {
  features {}

  skip_provider_registration = true
}

# Explanation: "features" is a nested block,
# but it appears before the non-block argument skip_provider_registration.
# That’s invalid for this rule:
# once we see a block, all subsequent items must also be blocks,
# not an attribute.

# We expect a single issue: "Argument 'skip_provider_registration' must not come after a nested block"
