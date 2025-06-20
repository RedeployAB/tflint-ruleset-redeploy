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

	// Meta argument names
	ArgDependsOn = "depends_on"
	ArgLifecycle = "lifecycle"
	ArgCount     = "count"
	ArgForEach   = "for_each"
	ArgProvider  = "provider"

	// For modules
	TypeModule = "module"

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
