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

	// Meta argument names
	ArgDependsOn = "depends_on"
	ArgLifecycle = "lifecycle"
	ArgCount     = "count"
	ArgForEach   = "for_each"
	ArgProvider  = "provider"

	// For modules
	TypeModule = "module"

	// Common string constants
	StringFalse = "false"
)
