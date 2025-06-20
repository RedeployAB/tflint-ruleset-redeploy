# Error Handling Improvements for Helper Functions

This document summarizes the error handling improvements made to the helper functions in the tflint-ruleset-redeploy project.

## Changes Made

### 1. Modified `GetAttributeRawText` Function

**File:** `rules/helper.go`

**Changes:**
- Changed function signature from `func GetAttributeRawText(attr *hclsyntax.Attribute, fileBytes []byte) string` to `func GetAttributeRawText(attr *hclsyntax.Attribute, fileBytes []byte) (string, error)`
- Added comprehensive input validation:
  - Check for nil attribute
  - Check for nil attribute expression
  - Check for nil fileBytes
  - Check for negative start byte
  - Enhanced bounds checking with descriptive error messages

**Error Cases Handled:**
- `attribute is nil`
- `attribute expression is nil`
- `fileBytes is nil`
- `expression start byte (%d) is negative`
- `expression end byte (%d) exceeds file length (%d)`
- `invalid range: start byte (%d) >= end byte (%d)`

### 2. Updated All Calling Sites

**Files Modified:**
- `rules/terraform_variable_sensitive.go`
- `rules/terraform_variable_ephemeral.go`
- `rules/terraform_output_sensitive.go`
- `rules/terraform_output_ephemeral.go`
- `rules/terraform_variable_nullable.go` (3 helper functions)

**Pattern Applied:**
```go
// Before:
src := GetAttributeRawText(attr, fileBytes)

// After:
src, err := GetAttributeRawText(attr, fileBytes)
if err != nil {
    // If we can't parse the attribute text, skip this check
    return nil
}
```

### 3. Enhanced `terraform_variable_nullable.go` Helper Functions

**Functions Updated:**
- `isTypeBool(attr *hclsyntax.Attribute, fileBytes []byte) (bool, error)`
- `isAttrNull(attr *hclsyntax.Attribute, fileBytes []byte) (bool, error)`
- `isAttrTrue(attr *hclsyntax.Attribute, fileBytes []byte) (bool, error)`

All now properly handle and propagate errors from `GetAttributeRawText`.

### 4. Improved `ExprAsKeyword` Error Handling

**Files Modified:**
- `rules/terraform_provider_minimum_major_version.go`
- `rules/terraform_provider_source.go`

**Changes:**
- Added explicit checks for empty returns from `hcl.ExprAsKeyword()`
- Skip processing when key extraction fails instead of silently continuing with invalid data

**Pattern Applied:**
```go
// Before:
if strings.TrimSpace(hcl.ExprAsKeyword(kv.KeyExpr)) != argVersion {
    continue
}

// After:
keyName := strings.TrimSpace(hcl.ExprAsKeyword(kv.KeyExpr))
if keyName == "" {
    // Skip if key extraction fails
    continue
}
if keyName != argVersion {
    continue
}
```

### 5. Enhanced Test Coverage

**File:** `rules/helper_test.go`

**Additions:**
- Updated existing tests to handle the new error return from `GetAttributeRawText`
- Added new test cases for error conditions:
  - Test for nil attribute parameter
  - Test for nil fileBytes parameter
- All tests verify that errors are properly returned and handled

## Error Handling Strategy

The error handling strategy follows these principles:

1. **Fail Fast**: Input validation happens early with descriptive error messages
2. **Graceful Degradation**: When attribute parsing fails, rules skip the check rather than crash
3. **Comprehensive Bounds Checking**: All array/slice access is properly validated
4. **Clear Error Messages**: All error messages include context and values for debugging
5. **Backward Compatibility**: Changes maintain existing behavior while adding safety

## Testing

All changes have been verified with:
- **Unit Tests**: All existing tests pass with the new error handling
- **Integration Tests**: Build and compilation successful
- **Error Path Testing**: New test cases verify error conditions are properly handled

## Impact

These improvements provide:
- **Better Reliability**: Prevents crashes from invalid input
- **Improved Debugging**: Clear error messages help identify issues
- **Safer Operations**: Comprehensive bounds checking prevents memory access errors
- **Maintainability**: Consistent error handling patterns across the codebase

The changes are backward compatible and don't affect the external API of the rules.