# Rules

Redeploy's Terraform rules implement recommendations and enforce best practices as per the [Terraform documentation](https://www.terraform.io/language). This ruleset provides a comprehensive set of rules to ensure consistent Terraform code style and usage across your projects.

All rules are enabled by default. For a detailed description of each rule, see the documentation below.

| Rule                                                         | Description                                                                             | Severity |
| ------------------------------------------------------------ | --------------------------------------------------------------------------------------- | -------- |
| [terraform_basic_module_structure](terraform_basic_module_structure.md)   | Ensures modules follow a basic Terraform module structure.                              | Warning  |
| [terraform_block_format](terraform_block_format.md)                         | Enforces consistent formatting of Terraform blocks.                                     | Error    |
| [terraform_block_order](terraform_block_order.md)                           | Validates the order in which blocks appear.                                             | Error    |
| [terraform_config_block](terraform_config_block.md)                         | Checks that configuration blocks are declared in the proper file format.                | Error    |
| [terraform_filename_convention](terraform_filename_convention.md)           | Enforces naming conventions for Terraform files.                                        | Error    |
| [terraform_locals_file](terraform_locals_file.md)                           | Validates the naming and structure of locals files.                                     | Error    |
| [terraform_locals_mirror_assignment](terraform_locals_mirror_assignment.md) | Ensures locals assignments mirror expected patterns.                                    | Error    |
| [terraform_meta_argument_format](terraform_meta_argument_format.md)         | Enforces formatting of meta arguments within resource and module blocks.                | Error    |
| [terraform_meta_argument_order](terraform_meta_argument_order.md)           | Validates the order of meta arguments.                                                  | Error    |
| [terraform_module_depends_on](terraform_module_depends_on.md)               | Disallows the use of `depends_on` in module blocks.                                     | Warning  |
| [terraform_no_leading_trailing_blank_lines](terraform_no_leading_trailing_blank_lines.md) | Ensures Terraform files do not have extra leading or trailing blank lines. | Error    |
| [terraform_output_argument_order](terraform_output_argument_order.md)       | Checks that output block arguments are ordered correctly.                               | Error    |
| [terraform_output_ephemeral](terraform_output_ephemeral.md)                 | Validates outputs that reference ephemeral resources.                                   | Error    |
| [terraform_output_file](terraform_output_file.md)                           | Ensures that output file declarations follow a prescribed structure.                    | Error    |
| [terraform_output_order](terraform_output_order.md)                         | Enforces a specific ordering for output blocks.                                         | Error    |
| [terraform_output_sensitive](terraform_output_sensitive.md)                 | Checks that outputs marked as sensitive are handled appropriately.                      | Error    |
| [terraform_provider_minimum_major_version](terraform_provider_minimum_major_version.md) | Ensures providers meet minimum major version constraints.               | Error    |
| [terraform_provider_source](terraform_provider_source.md)                   | Validates the ordering of provider source declarations.                                 | Error    |
| [terraform_resource_argument_order](terraform_resource_argument_order.md)   | Checks that resource arguments are in the correct order.                                | Error    |
| [terraform_resource_name](terraform_resource_name.md)                       | Enforces naming conventions for Terraform resources.                                    | Error    |
| [terraform_single_blank_lines](terraform_single_blank_lines.md)             | Disallows consecutive blank lines within files.                                         | Error    |
| [terraform_source_format](terraform_source_format.md)                       | Enforces proper formatting for module source declarations.                              | Error    |
| [terraform_tags_argument](terraform_tags_argument.md)                       | Validates the format of tag arguments in resource blocks.                               | Error    |
| [terraform_variable_argument_order](terraform_variable_argument_order.md)   | Checks that variable declarations list their arguments in the correct order.            | Error    |
| [terraform_variable_ephemeral](terraform_variable_ephemeral.md)             | Validates ephemeral variable declarations.                                              | Error    |
| [terraform_variable_file](terraform_variable_file.md)                       | Ensures variable files follow the required structure.                                   | Error    |
| [terraform_variable_nullable](terraform_variable_nullable.md)               | Enforces proper handling of nullable variable declarations.                             | Error    |
| [terraform_variable_order](terraform_variable_order.md)                     | Enforces the correct ordering of variables: required variables first (alphabetical), then optional variables (alphabetical). | Error    |
| [terraform_variable_sensitive](terraform_variable_sensitive.md)             | Validates that sensitive variables are declared correctly.                              | Error    |
