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

	// Type constants
	TypeBool   = "bool"
	TypeString = "string"
	TypeNumber = "number"
	TypeAny    = "any"
)
