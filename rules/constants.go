package rules

const (
	// Block types
	TypeResource  = "resource"
	TypeData      = "data"
	TypeTerraform = "terraform"
	TypeProvider  = "provider"
	TypeAttr      = "attr"
	TypeBlock     = "block"
	TypeVariable  = "variable"
	TypeLocals    = "locals"
	TypeOutput    = "output"

	// Sub-block types
	TypeRequiredProviders = "required_providers"
	TypePrecondition      = "precondition"
	TypeValidation        = "validation"

	// Meta argument names
	ArgDependsOn = "depends_on"
	ArgLifecycle = "lifecycle"
	ArgCount     = "count"
	ArgForEach   = "for_each"
	ArgProvider  = "provider"

	// Common block/output/variable argument and label names
	ArgDescription = "description"
	ArgType        = "type"
	ArgDefault     = "default"
	ArgEphemeral   = "ephemeral"
	ArgSensitive   = "sensitive"
	ArgNullable    = "nullable"
	ArgName        = "name"

	// Conventional module file names
	FileMain      = "main.tf"
	FileVariables = "variables.tf"
	FileOutputs   = "outputs.tf"
	FileTerraform = "terraform.tf"

	// For modules
	TypeModule = "module"

	// Lifecycle/operational block types (Terraform 1.1+)
	TypeMoved   = "moved"
	TypeImport  = "import"
	TypeRemoved = "removed"
	TypeCheck   = "check"

	// Reference types
	TypeVar   = "var"
	TypeLocal = "local"

	// Common string constants
	StringFalse = "false"
	StringTrue  = "true"
	StringNull  = "null"

	// Type constants
	TypeBool   = "bool"
	TypeString = "string"
	TypeNumber = "number"
	TypeAny    = "any"
)
