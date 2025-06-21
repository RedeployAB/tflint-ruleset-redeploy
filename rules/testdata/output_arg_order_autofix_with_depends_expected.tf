output "with_depends" {
  description = "Instance ID"
  value       = var.instance_id

  depends_on = [aws_instance.example]
}
