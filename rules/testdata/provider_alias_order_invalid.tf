provider "aws" {
  region = "us-east-1"
  alias  = "us_east_1"
}

provider "google" {
  alias   = "europe"
  project = "example"
}

provider "google" {
  project = "example"
}
