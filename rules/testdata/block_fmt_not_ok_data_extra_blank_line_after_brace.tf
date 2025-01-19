data "aws_ami" "example" {

  filter {
    name   = "xyz"
    values = ["abc"]
  }
}
