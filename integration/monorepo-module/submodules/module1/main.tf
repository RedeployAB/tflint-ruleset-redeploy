# Submodule resource
resource "null_resource" "example" {
  triggers = {
    value = "submodule"
  }
}
