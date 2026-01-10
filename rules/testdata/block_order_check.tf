terraform {}

check "health_check" {
  data "http" "example" {
    url = "https://example.com"
  }

  assert {
    condition     = data.http.example.status_code == 200
    error_message = "Health check failed"
  }
}

resource "aws_instance" "example" {}
