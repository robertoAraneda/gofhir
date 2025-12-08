package fhirpath

import (
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/funcs"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/parser/grammar"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// Expression represents a compiled FHIRPath expression.
type Expression struct {
	source string
	tree   *grammar.EntireExpressionContext
}

// Evaluate executes the expression against a JSON resource.
func (e *Expression) Evaluate(resource []byte) (types.Collection, error) {
	ctx := eval.NewContext(resource)
	return e.EvaluateWithContext(ctx)
}

// EvaluateWithContext executes the expression with a custom context.
func (e *Expression) EvaluateWithContext(ctx *eval.Context) (types.Collection, error) {
	evaluator := eval.NewEvaluator(ctx, funcs.GetRegistry())
	return evaluator.Evaluate(e.tree)
}

// String returns the original expression string.
func (e *Expression) String() string {
	return e.source
}
