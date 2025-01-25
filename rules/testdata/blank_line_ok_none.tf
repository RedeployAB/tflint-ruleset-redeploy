resource "random_uuid" "role_assignment" {
  lifecycle {
    replace_triggered_by = [
      null_resource.role_assignment[each.key]
    ]
  }
}
