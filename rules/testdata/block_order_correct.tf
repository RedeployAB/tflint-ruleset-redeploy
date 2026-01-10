terraform {
  required_version = ">= 1.0.0"
}

provider "aws" {}

data "aws_iam_user" "example" {}

resource "aws_instance" "example" {}
