resource "random_uuid" "role_assignment" {
  for_each = local.role_assignments

  lifecycle {
    replace_triggered_by = [
      null_resource.role_assignment[each.key]
    ]
  }
}
