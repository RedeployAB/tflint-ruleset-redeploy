+terraform {
+  required_version = "~> 1.7.0"
+
+  required_providers {
+    azurerm = {
+      source  = "hashicorp/azurerm"
+      version = ">= 3.97.0, < 4.0.0"
+    }
+  }
+}
+
+provider "azurerm" {
+  features {}
+}
+
+resource "azurerm_resource_group" "example" {
+  name     = var.resource_group_name
+  location = var.location
+  tags     = var.tags
+}
