resource "aws_instance" "this" {
  lifecycle {
    ignore_changes = [tags, ami]
  }
}

resource "aws_instance" "single" {
  lifecycle {
    ignore_changes = [tags]
  }
}

# `all` inside a list is an attribute reference, not the bare keyword.
resource "aws_instance" "all_in_list" {
  lifecycle {
    ignore_changes = [all]
  }
}
