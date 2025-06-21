# Autofix Development Plan

This document outlines the plan for implementing autofix functionality across the TFLint ruleset. Each rule is categorized by implementation difficulty and includes notes for developers.

## Implementation Guide

Before starting, please read the [Autofix Implementation Guide](docs/developer/autofix-guide.md) for detailed instructions and examples.

## Priority 1: Easy Rules (Good for First Contributors)

These rules require simple text manipulation and are excellent starting points:

### ✅ Completed
- [x] **terraform_variable_sensitive** - Remove `sensitive = false` attributes
  - Fixed to properly remove entire line including newline
  - Added comprehensive autofix tests

- [x] **terraform_output_sensitive** - Remove `sensitive = false` attributes
  - Implemented same fix as variable_sensitive (removes entire line)
  - Added comprehensive autofix tests including edge cases

- [x] **terraform_variable_nullable** - Remove unnecessary `nullable` declarations
  - Remove `nullable = true` (since true is the default)
  - Remove `nullable` when `default = null` (nullable not needed)
  - Added comprehensive autofix tests for both scenarios

### 🔲 To Do

- [ ] **terraform_locals_mirror_assignment** - Replace indirect reference with direct
  - Simple `ReplaceText()` on the expression
  - Example: `local.foo` → `var.original_value`

## Priority 2: Formatting Rules

These rules fix spacing and formatting issues:

- [ ] **terraform_single_blank_lines** - Replace multiple blank lines with single
  - Use `ReplaceText()` to fix spacing
  - Need to handle edge cases at file boundaries

- [ ] **terraform_no_leading_trailing_blank_lines** - Remove blank lines at file start/end
  - Use `Remove()` for the blank line ranges

- [ ] **terraform_source_format** - Fix module source block formatting
  - Ensure proper blank lines between source/version and other attributes
  - Use combination of `InsertTextBefore()` and `Remove()`

## Priority 3: Ordering Rules (Medium Complexity)

These rules require collecting multiple blocks and reordering them:

- [ ] **terraform_variable_order** - Reorder variable blocks
  - Collect all variables in a file
  - Sort: required first (alphabetical), then optional (alphabetical)
  - Use `TextAt()` to preserve formatting
  - Replace entire range with sorted content

- [ ] **terraform_output_order** - Reorder output blocks
  - Similar to variable_order
  - Sort alphabetically by output name

- [ ] **terraform_resource_argument_order** - Reorder resource arguments
  - Arguments before blocks
  - Tags at the end of arguments
  - Meta-arguments after all other content

- [ ] **terraform_meta_argument_order** - Order meta-arguments correctly
  - Fixed order: count/for_each, provider, lifecycle, depends_on

- [ ] **terraform_variable_argument_order** - Order variable block arguments
  - Fixed order: description, type, default, sensitive, nullable, validation

- [ ] **terraform_output_argument_order** - Order output block arguments
  - Fixed order: description, value, sensitive, depends_on, precondition

## Priority 4: Complex Rules

These rules require significant logic or file manipulation:

- [ ] **terraform_block_order** - Reorder all blocks in a file
  - Very complex - needs to handle all block types
  - Preserve comments and formatting
  - Consider making this opt-in only

- [ ] **terraform_basic_module_structure** - Move blocks to correct files
  - Extremely complex - requires file creation/deletion
  - May not be suitable for autofix
  - Consider documenting as "manual fix only"

## Rules Not Suitable for Autofix

These rules require human judgment or are too complex:

- **terraform_filename_convention** - Would require file renaming
- **terraform_resource_name** - Changing resource names breaks references
- **terraform_module_depends_on** - Requires understanding of dependencies
- **terraform_tags_argument** - Content is project-specific
- **terraform_provider_minimum_major_version** - Requires compatibility analysis

## Development Tips

1. Start with Priority 1 rules to get familiar with the API
2. Test your fixes with various edge cases
3. Always preserve formatting using `TextAt()` when moving content
4. Handle JSON syntax files by returning `tflint.ErrFixNotSupported`
5. Write tests for your autofix implementations

## Assignment Tracking

| Rule | Assigned To | Status | PR |
|------|------------|--------|-----|
| terraform_variable_sensitive | @lars | ✅ Complete | - |
| terraform_output_sensitive | @lars | ✅ Complete | - |
| terraform_variable_nullable | @assistant | ✅ Complete | - |
| terraform_locals_mirror_assignment | - | 🔲 Available | - |
| terraform_single_blank_lines | - | 🔲 Available | - |
| terraform_variable_order | @lars | 🚧 Started | - |

---

**Note**: This is a living document. Please update the assignment tracking table when you start working on a rule.