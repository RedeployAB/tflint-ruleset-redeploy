output "public" {
  value = "public value"
}

output "secret" {
  value     = "secret value"
  sensitive = true
}
