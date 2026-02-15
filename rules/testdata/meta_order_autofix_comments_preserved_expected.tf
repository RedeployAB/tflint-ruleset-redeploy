resource "azurerm_container_app" "main" {
  name                         = "ca-example"
  resource_group_name          = "rg-example"
  container_app_environment_id = azurerm_container_app_environment.main.id
  revision_mode                = "Single"

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.app.id]
  }

  # Registry configuration for GHCR
  registry {
    server   = "ghcr.io"
    username = var.ghcr_username
    identity = azurerm_user_assigned_identity.app.id
  }

  # Secrets from Key Vault
  secret {
    name                = "ghcr-pat"
    identity            = azurerm_user_assigned_identity.app.id
    key_vault_secret_id = "${azurerm_key_vault.main.vault_uri}secrets/ghcr-pat"
  }

  secret {
    name                = "slack-bot-token"
    identity            = azurerm_user_assigned_identity.app.id
    key_vault_secret_id = "${azurerm_key_vault.main.vault_uri}secrets/slack-bot-token"
  }

  ingress {
    external_enabled = true
    target_port      = 3000

    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  template {
    min_replicas = 1
    max_replicas = 3

    container {
      name   = "app"
      image  = "ghcr.io/example/app:latest"
      cpu    = 0.25
      memory = "0.5Gi"

      env {
        name        = "SLACK_BOT_TOKEN"
        secret_name = "slack-bot-token"
      }
    }
  }

  tags = local.default_tags

  lifecycle {
    ignore_changes = [
      # Image is updated by release workflow
      template[0].container[0].image,
      # App configuration env vars are set by release workflow
      template[0].container[0].env,
    ]
  }

  depends_on = [
    azurerm_role_assignment.app_keyvault
  ]
}
