package rules

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformSourceFormatRule enforces a blank line after the last of "source" or "version"
// within a module block, but ONLY if additional arguments follow. If the block ends, no extra
// blank line is required.
//
// Examples:
//  module "example" {
//    source  = "something"  # OK if block ends here
//  }
//
//  module "example" {
//    source  = "something"
//    version = "x.x.x"      # OK if block ends here
//  }
//
//  module "example" {
//    source  = "something"
//    version = "x.x.x"
//
//    property = "value"     # OK because there's a blank line
//  }
//
//  module "example" {
//    source  = "something"
//    property = "value"     # OK because there's a blank line
//  }
//
//  module "example" {
//    source  = "something"
//    version = "x.x.x"
//    property = "value"     # NOT OK, missing a blank line after version
//  }

type TerraformSourceFormatRule struct {
	tflint.DefaultRule
}

func NewTerraformSourceFormatRule() *TerraformSourceFormatRule {
	return &TerraformSourceFormatRule{}
}

func (r *TerraformSourceFormatRule) Name() string {
	return "terraform_source_format"
}

func (r *TerraformSourceFormatRule) Enabled() bool {
	return true
}

func (r *TerraformSourceFormatRule) Severity() tflint.Severity {
	return tflint.ERROR
}

func (r *TerraformSourceFormatRule) Link() string {
	return ""
}

func (r *TerraformSourceFormatRule) Check(runner tflint.Runner) error {
	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	for filename, file := range files {
		// Skip empty or nil
		if file == nil || file.Bytes == nil {
			continue
		}

		syntaxFile, diags := hclsyntax.ParseConfig(file.Bytes, filename, hcl.InitialPos)
		if diags.HasErrors() {
			continue
		}

		if body, ok := syntaxFile.Body.(*hclsyntax.Body); ok {
			if err := r.processBody(body, filename, runner); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *TerraformSourceFormatRule) processBody(body *hclsyntax.Body, filename string, runner tflint.Runner) error {
	for _, block := range body.Blocks {
		// Only check "module" blocks
		if block.Type == "module" {
			if err := r.checkModuleBlock(block, filename, runner); err != nil {
				return err
			}
		}
		// Recurse into nested blocks
		if err := r.processBody(block.Body, filename, runner); err != nil {
			return err
		}
	}
	return nil
}

func (r *TerraformSourceFormatRule) checkModuleBlock(block *hclsyntax.Block, filename string, runner tflint.Runner) error {
	// We'll parse the "source" and "version" lines by scanning the block's actual text lines.
	// Then if there's a "source" or "version," check whether we must require a blank line after it.
	srcRange := block.Body.Range()

	files, err := runner.GetFiles()
	if err != nil {
		return err
	}

	f, ok := files[srcRange.Filename]
	if !ok || f.Bytes == nil {
		return nil
	}

	lines := strings.Split(string(f.Bytes), "\n")

	startLine := srcRange.Start.Line - 1
	endLine := srcRange.End.Line - 1
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	// Identify line numbers for "source" and "version".
	// We'll store whichever is last (max line #).
	sourceLine := -1
	versionLine := -1

	for l := startLine; l <= endLine && l < len(lines); l++ {
		text := strings.TrimSpace(lines[l])
		// Skip empty lines (which have no text)
		if text == "" {
			continue
		}
		// We want to see if this line starts with "source " or "source="
		if strings.HasPrefix(text, "source ") || strings.HasPrefix(text, "source=") {
			sourceLine = l
		}
		if strings.HasPrefix(text, "version ") || strings.HasPrefix(text, "version=") {
			versionLine = l
		}
	}

	// If neither found, no issues
	if sourceLine < 0 && versionLine < 0 {
		return nil
	}

	// We only require a blank line if there's something *after* the last one
	lastOfTheTwo := max(sourceLine, versionLine)
	if lastOfTheTwo < 0 {
		return nil
	}

	// The next line after lastOfTheTwo
	nextLineIdx := lastOfTheTwo + 1
	// If nextLineIdx > endLine, then the block ends immediately, no blank line required
	if nextLineIdx > endLine {
		return nil
	}

	for nextLineIdx <= endLine {
		nextText := strings.TrimSpace(lines[nextLineIdx])
		if nextText == "" {
			// Found a blank line => good => done
			return nil
		} else if strings.HasPrefix(nextText, "//") || strings.HasPrefix(nextText, "#") {
			// Skip comment lines
			nextLineIdx++
			continue
		} else if nextText == "}" {
			// Next line is the closing brace => no blank line required
			return nil
		} else {
			// Next line is not blank, not a comment, not the closing brace => we must error
			rng := hcl.Range{
				Filename: srcRange.Filename,
				Start:    hcl.Pos{Line: nextLineIdx + 1, Column: 1},
				End:      hcl.Pos{Line: nextLineIdx + 1, Column: 1},
			}
			return runner.EmitIssue(
				r,
				fmt.Sprintf("Expected a blank line after '%s'", pickAttrName(sourceLine, versionLine, lastOfTheTwo)),
				rng,
			)
		}
	}

	return nil
}

func pickAttrName(srcLine, verLine, last int) string {
	switch last {
	case srcLine:
		return "source"
	case verLine:
		return "version"
	default:
		return "source"
	}
}
