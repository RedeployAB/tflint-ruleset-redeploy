output "with_depends" {
  depends_on  = [aws_instance.example]
  value       = var.instance_id
  description = "Instance ID"
}
