output "public" {
  value     = "public value"
  sensitive = false
}

output "secret" {
  value     = "secret value"
  sensitive = true
}
