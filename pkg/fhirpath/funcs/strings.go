package funcs

import (
	"strings"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func init() {
	// Register string functions
	Register(FuncDef{
		Name:    "startsWith",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnStartsWith,
	})

	Register(FuncDef{
		Name:    "endsWith",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnEndsWith,
	})

	Register(FuncDef{
		Name:    "contains",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnContains,
	})

	Register(FuncDef{
		Name:    "replace",
		MinArgs: 2,
		MaxArgs: 2,
		Fn:      fnReplace,
	})

	Register(FuncDef{
		Name:    "matches",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnMatches,
	})

	Register(FuncDef{
		Name:    "replaceMatches",
		MinArgs: 2,
		MaxArgs: 2,
		Fn:      fnReplaceMatches,
	})

	Register(FuncDef{
		Name:    "indexOf",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnIndexOf,
	})

	Register(FuncDef{
		Name:    "substring",
		MinArgs: 1,
		MaxArgs: 2,
		Fn:      fnSubstring,
	})

	Register(FuncDef{
		Name:    "lower",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnLower,
	})

	Register(FuncDef{
		Name:    "upper",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnUpper,
	})

	Register(FuncDef{
		Name:    "toChars",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToChars,
	})

	Register(FuncDef{
		Name:    "split",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnSplit,
	})

	Register(FuncDef{
		Name:    "join",
		MinArgs: 0,
		MaxArgs: 1,
		Fn:      fnJoin,
	})

	Register(FuncDef{
		Name:    "trim",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnTrim,
	})

	Register(FuncDef{
		Name:    "length",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnLength,
	})
}

// fnStartsWith returns true if the string starts with the given prefix.
func fnStartsWith(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	prefix, ok := toStringArg(args[0])
	if !ok {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewBoolean(strings.HasPrefix(str, prefix))}, nil
}

// fnEndsWith returns true if the string ends with the given suffix.
func fnEndsWith(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	suffix, ok := toStringArg(args[0])
	if !ok {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewBoolean(strings.HasSuffix(str, suffix))}, nil
}

// fnContains returns true if the string contains the given substring.
func fnContains(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	substr, ok := toStringArg(args[0])
	if !ok {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewBoolean(strings.Contains(str, substr))}, nil
}

// fnReplace replaces all occurrences of pattern with substitution.
func fnReplace(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	pattern, ok := toStringArg(args[0])
	if !ok {
		return types.Collection{}, nil
	}

	substitution, ok := toStringArg(args[1])
	if !ok {
		return types.Collection{}, nil
	}

	result := strings.ReplaceAll(str, pattern, substitution)
	return types.Collection{types.NewString(result)}, nil
}

// fnMatches returns true if the string matches the regex pattern.
// Uses cached regex compilation with ReDoS protection.
func fnMatches(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	pattern, ok := toStringArg(args[0])
	if !ok {
		return types.Collection{}, nil
	}

	// Use regex cache with timeout protection
	matched, err := DefaultRegexCache.MatchWithTimeout(ctx.Context(), pattern, str)
	if err != nil {
		return nil, err
	}

	return types.Collection{types.NewBoolean(matched)}, nil
}

// fnReplaceMatches replaces regex matches with substitution.
// Uses cached regex compilation with ReDoS protection.
func fnReplaceMatches(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	pattern, ok := toStringArg(args[0])
	if !ok {
		return types.Collection{}, nil
	}

	substitution, ok := toStringArg(args[1])
	if !ok {
		return types.Collection{}, nil
	}

	// Use regex cache with timeout protection
	result, err := DefaultRegexCache.ReplaceWithTimeout(ctx.Context(), pattern, str, substitution)
	if err != nil {
		return nil, err
	}

	return types.Collection{types.NewString(result)}, nil
}

// fnIndexOf returns the index of the first occurrence of substring.
func fnIndexOf(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	substr, ok := toStringArg(args[0])
	if !ok {
		return types.Collection{}, nil
	}

	idx := strings.Index(str, substr)
	return types.Collection{types.NewInteger(int64(idx))}, nil
}

// fnSubstring returns a substring starting at the given index.
func fnSubstring(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	start, err := toInteger(args[0])
	if err != nil {
		return nil, err
	}

	if start < 0 || int(start) >= len(str) {
		return types.Collection{}, nil
	}

	// Optional length parameter
	if len(args) > 1 {
		length, err := toInteger(args[1])
		if err != nil {
			return nil, err
		}
		end := int(start + length)
		if end > len(str) {
			end = len(str)
		}
		return types.Collection{types.NewString(str[start:end])}, nil
	}

	return types.Collection{types.NewString(str[start:])}, nil
}

// fnLower converts string to lowercase.
func fnLower(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewString(strings.ToLower(str))}, nil
}

// fnUpper converts string to uppercase.
func fnUpper(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewString(strings.ToUpper(str))}, nil
}

// fnToChars converts string to a collection of single characters.
func fnToChars(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	result := types.Collection{}
	for _, ch := range str {
		result = append(result, types.NewString(string(ch)))
	}

	return result, nil
}

// fnSplit splits a string by the given separator.
func fnSplit(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	separator, ok := toStringArg(args[0])
	if !ok {
		return types.Collection{}, nil
	}

	parts := strings.Split(str, separator)
	result := types.Collection{}
	for _, part := range parts {
		result = append(result, types.NewString(part))
	}

	return result, nil
}

// fnJoin joins a collection of strings with an optional separator.
func fnJoin(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewString("")}, nil
	}

	separator := ""
	if len(args) > 0 {
		if sep, ok := toStringArg(args[0]); ok {
			separator = sep
		}
	}

	parts := make([]string, 0, len(input))
	for _, item := range input {
		if s, ok := item.(types.String); ok {
			parts = append(parts, s.Value())
		} else {
			parts = append(parts, item.String())
		}
	}

	return types.Collection{types.NewString(strings.Join(parts, separator))}, nil
}

// fnTrim removes leading and trailing whitespace.
func fnTrim(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewString(strings.TrimSpace(str))}, nil
}

// fnLength returns the length of the string.
func fnLength(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	str, ok := toString(input)
	if !ok {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewInteger(int64(len(str)))}, nil
}

// Helper functions

// toString extracts a string from a collection's first element.
func toString(col types.Collection) (string, bool) {
	if col.Empty() {
		return "", false
	}
	if s, ok := col[0].(types.String); ok {
		return s.Value(), true
	}
	return col[0].String(), true
}

// toStringArg extracts a string from an argument.
func toStringArg(arg interface{}) (string, bool) {
	switch v := arg.(type) {
	case types.Collection:
		return toString(v)
	case types.String:
		return v.Value(), true
	case string:
		return v, true
	default:
		return "", false
	}
}
