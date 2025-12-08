package fhirpath

import (
	"context"
	"time"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// EvalOptions configures expression evaluation.
type EvalOptions struct {
	// Context for cancellation and timeout
	Ctx context.Context

	// Timeout for evaluation (0 means no timeout)
	Timeout time.Duration

	// MaxDepth limits recursion depth for descendants() (0 means default of 100)
	MaxDepth int

	// MaxCollectionSize limits output collection size (0 means no limit)
	MaxCollectionSize int

	// Variables are external variables accessible via %name
	Variables map[string]types.Collection

	// Resolver handles reference resolution for resolve() function
	Resolver ReferenceResolver
}

// DefaultOptions returns default evaluation options suitable for production.
func DefaultOptions() *EvalOptions {
	return &EvalOptions{
		Ctx:               context.Background(),
		Timeout:           5 * time.Second,
		MaxDepth:          100,
		MaxCollectionSize: 10000,
		Variables:         make(map[string]types.Collection),
	}
}

// EvalOption is a functional option for configuring evaluation.
type EvalOption func(*EvalOptions)

// WithContext sets the context for cancellation.
func WithContext(ctx context.Context) EvalOption {
	return func(o *EvalOptions) {
		o.Ctx = ctx
	}
}

// WithTimeout sets the evaluation timeout.
func WithTimeout(d time.Duration) EvalOption {
	return func(o *EvalOptions) {
		o.Timeout = d
	}
}

// WithMaxDepth sets the maximum recursion depth.
func WithMaxDepth(depth int) EvalOption {
	return func(o *EvalOptions) {
		o.MaxDepth = depth
	}
}

// WithMaxCollectionSize sets the maximum output collection size.
func WithMaxCollectionSize(size int) EvalOption {
	return func(o *EvalOptions) {
		o.MaxCollectionSize = size
	}
}

// WithVariable sets an external variable.
func WithVariable(name string, value types.Collection) EvalOption {
	return func(o *EvalOptions) {
		if o.Variables == nil {
			o.Variables = make(map[string]types.Collection)
		}
		o.Variables[name] = value
	}
}

// WithResolver sets the reference resolver.
func WithResolver(r ReferenceResolver) EvalOption {
	return func(o *EvalOptions) {
		o.Resolver = r
	}
}

// ReferenceResolver resolves FHIR references for the resolve() function.
type ReferenceResolver interface {
	// Resolve takes a reference string (e.g., "Patient/123") and returns the resource.
	Resolve(ctx context.Context, reference string) ([]byte, error)
}

// EvaluateWithOptions evaluates an expression with custom options.
func (e *Expression) EvaluateWithOptions(resource []byte, opts ...EvalOption) (types.Collection, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Create context with timeout if specified
	ctx := options.Ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Create evaluation context
	evalCtx := eval.NewContext(resource)

	// Set variables
	for name, value := range options.Variables {
		evalCtx.SetVariable(name, value)
	}

	// Set limits in context
	evalCtx.SetLimit("maxDepth", options.MaxDepth)
	evalCtx.SetLimit("maxCollectionSize", options.MaxCollectionSize)
	evalCtx.SetContext(ctx)

	// Set resolver if provided
	if options.Resolver != nil {
		evalCtx.SetResolver(newResolverAdapter(options.Resolver))
	}

	return e.EvaluateWithContext(evalCtx)
}

// resolverAdapter adapts ReferenceResolver to eval.Resolver
type resolverAdapter struct {
	resolver ReferenceResolver
}

func newResolverAdapter(r ReferenceResolver) *resolverAdapter {
	return &resolverAdapter{resolver: r}
}

func (a *resolverAdapter) Resolve(ctx context.Context, reference string) ([]byte, error) {
	return a.resolver.Resolve(ctx, reference)
}
