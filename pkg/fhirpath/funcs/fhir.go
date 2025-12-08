package funcs

import (
	"strings"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func init() {
	// Register FHIR-specific functions
	Register(FuncDef{
		Name:    "resolve",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnResolve,
	})

	Register(FuncDef{
		Name:    "extension",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnExtension,
	})

	Register(FuncDef{
		Name:    "hasExtension",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnHasExtension,
	})

	Register(FuncDef{
		Name:    "getExtensionValue",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnGetExtensionValue,
	})

	Register(FuncDef{
		Name:    "getReferenceKey",
		MinArgs: 0,
		MaxArgs: 1,
		Fn:      fnGetReferenceKey,
	})
}

// fnResolve resolves a FHIR reference to the referenced resource.
// This function requires a resolver to be set in the context.
func fnResolve(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	resolver := ctx.GetResolver()
	if resolver == nil {
		// Without a resolver, we can't resolve references
		// Return empty collection as per FHIRPath spec
		return types.Collection{}, nil
	}

	result := types.Collection{}

	for _, item := range input {
		var reference string

		switch v := item.(type) {
		case types.String:
			reference = v.Value()
		case *types.ObjectValue:
			// Try to get the 'reference' field from a Reference object
			if ref, ok := v.Get("reference"); ok {
				if refStr, ok := ref.(types.String); ok {
					reference = refStr.Value()
				}
			}
		}

		if reference == "" {
			continue
		}

		// Resolve the reference
		resourceJSON, err := resolver.Resolve(ctx.Context(), reference)
		if err != nil {
			// Skip references that can't be resolved
			continue
		}

		// Parse the resolved resource
		col, err := types.JSONToCollection(resourceJSON)
		if err != nil {
			continue
		}

		result = append(result, col...)
	}

	return result, nil
}

// fnExtension returns extensions matching the given URL.
func fnExtension(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() || len(args) == 0 {
		return types.Collection{}, nil
	}

	// Get the extension URL to search for
	var url string
	if col, ok := args[0].(types.Collection); ok && !col.Empty() {
		if str, ok := col[0].(types.String); ok {
			url = str.Value()
		}
	}

	if url == "" {
		return types.Collection{}, nil
	}

	result := types.Collection{}

	for _, item := range input {
		obj, ok := item.(*types.ObjectValue)
		if !ok {
			continue
		}

		// Get the extension array
		extensions := obj.GetCollection("extension")
		for _, ext := range extensions {
			extObj, ok := ext.(*types.ObjectValue)
			if !ok {
				continue
			}

			// Check if the URL matches
			if extURL, ok := extObj.Get("url"); ok {
				if urlStr, ok := extURL.(types.String); ok {
					if urlStr.Value() == url {
						result = append(result, extObj)
					}
				}
			}
		}
	}

	return result, nil
}

// fnHasExtension returns true if any input element has an extension with the given URL.
func fnHasExtension(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	extensions, err := fnExtension(ctx, input, args)
	if err != nil {
		return nil, err
	}

	return types.Collection{types.NewBoolean(!extensions.Empty())}, nil
}

// fnGetExtensionValue returns the value of extensions matching the given URL.
func fnGetExtensionValue(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	extensions, err := fnExtension(ctx, input, args)
	if err != nil {
		return nil, err
	}

	result := types.Collection{}

	for _, ext := range extensions {
		extObj, ok := ext.(*types.ObjectValue)
		if !ok {
			continue
		}

		// Look for value[x] fields
		valueFields := []string{
			"valueString", "valueBoolean", "valueInteger", "valueDecimal",
			"valueDate", "valueDateTime", "valueTime", "valueCode",
			"valueCoding", "valueCodeableConcept", "valueQuantity",
			"valueReference", "valueIdentifier", "valuePeriod",
			"valueRange", "valueRatio", "valueAttachment",
			"valueUri", "valueUrl", "valueCanonical",
		}

		for _, field := range valueFields {
			if val, ok := extObj.Get(field); ok {
				result = append(result, val)
				break
			}
		}
	}

	return result, nil
}

// fnGetReferenceKey extracts the resource type and ID from a reference.
// Returns a string in the format "ResourceType/id" or just "id" if no type prefix.
func fnGetReferenceKey(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	// Optional argument: specific part to extract ("type", "id", or default "key")
	part := "key"
	if len(args) > 0 {
		if col, ok := args[0].(types.Collection); ok && !col.Empty() {
			if str, ok := col[0].(types.String); ok {
				part = str.Value()
			}
		}
	}

	result := types.Collection{}

	for _, item := range input {
		var reference string

		switch v := item.(type) {
		case types.String:
			reference = v.Value()
		case *types.ObjectValue:
			if ref, ok := v.Get("reference"); ok {
				if refStr, ok := ref.(types.String); ok {
					reference = refStr.Value()
				}
			}
		}

		if reference == "" {
			continue
		}

		// Parse the reference
		// Remove any URL prefix (e.g., "http://example.org/fhir/Patient/123")
		if idx := strings.LastIndex(reference, "/"); idx > 0 {
			// Check if there's a resource type prefix before this
			beforeSlash := reference[:idx]
			if lastSlashBefore := strings.LastIndex(beforeSlash, "/"); lastSlashBefore >= 0 {
				reference = beforeSlash[lastSlashBefore+1:] + "/" + reference[idx+1:]
			}
		}

		switch part {
		case "type":
			if idx := strings.Index(reference, "/"); idx > 0 {
				result = append(result, types.NewString(reference[:idx]))
			}
		case "id":
			if idx := strings.LastIndex(reference, "/"); idx >= 0 {
				result = append(result, types.NewString(reference[idx+1:]))
			} else {
				result = append(result, types.NewString(reference))
			}
		default: // "key" or any other value
			result = append(result, types.NewString(reference))
		}
	}

	return result, nil
}
