resource "azurerm_consumption_budget_subscription" "this" {
  count = local.enable_budget ? 1 : 0

  amount     = local.budget_amount
  time_grain = "Monthly"

  # Budget warning threshold
  notification {
    enabled   = true
    threshold = var.budget_threshold_warning_percentage
    operator  = "GreaterThanOrEqualTo"

    contact_emails = local.contact_emails
    contact_roles  = ["Owner"]
  }
}
