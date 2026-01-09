package rules

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Benchmark tests for rule performance

// Sample Terraform configurations for benchmarks
// Note: smallTerraformConfig is available for future benchmark expansion
var _ = smallTerraformConfig // Prevent unused variable lint error

const smallTerraformConfig = `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

variable "instance_type" {
  description = "The type of EC2 instance"
  type        = string
  default     = "t3.micro"
}

resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = var.instance_type

  tags = {
    Name = "example"
  }
}

output "instance_id" {
  value = aws_instance.example.id
}
`

// Medium config with multiple resources
const mediumTerraformConfig = `
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "instance_count" {
  description = "Number of instances"
  type        = number
  default     = 2
}

variable "enable_monitoring" {
  description = "Enable detailed monitoring"
  type        = bool
  default     = false
}

locals {
  common_tags = {
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = local.common_tags
}

resource "aws_subnet" "public" {
  count      = var.instance_count
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.${count.index}.0/24"

  tags = local.common_tags
}

resource "aws_security_group" "web" {
  name   = "web-sg"
  vpc_id = aws_vpc.main.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}

resource "aws_instance" "web" {
  count         = var.instance_count
  ami           = "ami-12345678"
  instance_type = "t3.micro"
  subnet_id     = aws_subnet.public[count.index].id

  vpc_security_group_ids = [aws_security_group.web.id]

  monitoring = var.enable_monitoring

  tags = merge(local.common_tags, {
    Name = "web-${count.index}"
  })
}

output "vpc_id" {
  description = "The VPC ID"
  value       = aws_vpc.main.id
}

output "instance_ids" {
  description = "The instance IDs"
  value       = aws_instance.web[*].id
}

output "public_ips" {
  description = "The public IPs"
  value       = aws_instance.web[*].public_ip
}
`

// BenchmarkHCLParsing benchmarks raw HCL parsing performance
func BenchmarkHCLParsing(b *testing.B) {
	content := []byte(mediumTerraformConfig)

	b.Run("ParseConfig", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			//nolint:errcheck // Benchmark intentionally ignores errors
			hclsyntax.ParseConfig(content, "main.tf", hcl.InitialPos)
		}
	})
}

// BenchmarkStringSplit benchmarks string splitting which was a bottleneck
func BenchmarkStringSplit(b *testing.B) {
	content := mediumTerraformConfig

	b.Run("Split", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = strings.Split(content, "\n")
		}
	})

	b.Run("SplitAndCount", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			lines := strings.Split(content, "\n")
			_ = len(lines)
		}
	})
}

// BenchmarkBytePositionCalculation benchmarks the byte position calculation
// that was repeated in multiple rules
func BenchmarkBytePositionCalculation(b *testing.B) {
	content := mediumTerraformConfig
	lines := strings.Split(content, "\n")

	b.Run("LinearScan", func(b *testing.B) {
		targetLine := 50
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bytePos := 0
			for j := 0; j < targetLine && j < len(lines); j++ {
				bytePos += len(lines[j]) + 1
			}
			_ = bytePos
		}
	})

	b.Run("PrecomputedOffsets", func(b *testing.B) {
		// Pre-compute line offsets
		offsets := make([]int, len(lines)+1)
		pos := 0
		for i, line := range lines {
			offsets[i] = pos
			pos += len(line) + 1
		}
		offsets[len(lines)] = pos

		targetLine := 50
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if targetLine < len(offsets) {
				_ = offsets[targetLine]
			}
		}
	})
}

// BenchmarkTrimSpace benchmarks the trimspace operation used in blank line checks
func BenchmarkTrimSpace(b *testing.B) {
	lines := strings.Split(mediumTerraformConfig, "\n")

	b.Run("TrimSpace", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, line := range lines {
				_ = strings.TrimSpace(line)
			}
		}
	})

	b.Run("TrimSpaceWithBlankCheck", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				_ = trimmed == ""
			}
		}
	})
}

// BenchmarkMapLookup benchmarks map vs switch for attribute order lookups
func BenchmarkMapLookup(b *testing.B) {
	orderMap := map[string]int{
		"description": 0,
		"type":        1,
		"default":     2,
		"ephemeral":   3,
		"sensitive":   4,
		"nullable":    5,
		"validation":  6,
	}

	lookupSwitch := func(name string) int {
		switch name {
		case "description":
			return 0
		case "type":
			return 1
		case "default":
			return 2
		case "ephemeral":
			return 3
		case "sensitive":
			return 4
		case "nullable":
			return 5
		case "validation":
			return 6
		}
		return -1
	}

	testNames := []string{"description", "type", "default", "ephemeral", "sensitive", "nullable", "validation", "unknown"}

	b.Run("MapLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range testNames {
				_ = orderMap[name]
			}
		}
	})

	b.Run("SwitchLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, name := range testNames {
				_ = lookupSwitch(name)
			}
		}
	})
}

// BenchmarkBlockTypeCheck benchmarks block type checking
func BenchmarkBlockTypeCheck(b *testing.B) {
	types := []string{"resource", "data", "module", "variable", "output", "provider", "terraform", "locals"}

	b.Run("ToLowerAndCompare", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, t := range types {
				_ = strings.ToLower(t) == "resource"
			}
		}
	})

	b.Run("EqualFold", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, t := range types {
				_ = strings.EqualFold(t, "resource")
			}
		}
	})

	b.Run("DirectCompare", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, t := range types {
				_ = t == "resource" || t == "Resource" || t == "RESOURCE"
			}
		}
	})
}

// BenchmarkTraversalKey benchmarks traversal key generation
func BenchmarkTraversalKey(b *testing.B) {
	// Create a sample traversal
	trav := hcl.Traversal{
		hcl.TraverseRoot{Name: "aws_instance"},
		hcl.TraverseAttr{Name: "example"},
		hcl.TraverseAttr{Name: "id"},
	}

	b.Run("StringsJoin", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = traversalKey(trav)
		}
	})

	b.Run("StringBuilder", func(b *testing.B) {
		traversalKeyBuilder := func(trav hcl.Traversal) string {
			var sb strings.Builder
			for i, step := range trav {
				if i > 0 {
					sb.WriteByte('.')
				}
				switch s := step.(type) {
				case hcl.TraverseRoot:
					sb.WriteString(s.Name)
				case hcl.TraverseAttr:
					sb.WriteString(s.Name)
				case hcl.TraverseIndex:
					sb.WriteString("[idx]")
				case hcl.TraverseSplat:
					sb.WriteString("[*]")
				}
			}
			return sb.String()
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = traversalKeyBuilder(trav)
		}
	})
}
