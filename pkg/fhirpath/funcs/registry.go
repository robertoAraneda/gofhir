// Package funcs provides FHIRPath function implementations.
package funcs

import (
	"sync"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
)

// FuncDef is an alias for eval.FuncDef.
type FuncDef = eval.FuncDef

// Registry holds registered functions.
type Registry struct {
	funcs map[string]eval.FuncDef
	mu    sync.RWMutex
}

// globalRegistry is the default function registry.
var globalRegistry = NewRegistry()

// NewRegistry creates a new function registry.
func NewRegistry() *Registry {
	r := &Registry{
		funcs: make(map[string]eval.FuncDef),
	}
	return r
}

// Register adds a function to the registry.
func (r *Registry) Register(def eval.FuncDef) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.funcs[def.Name] = def
}

// Get retrieves a function by name.
func (r *Registry) Get(name string) (eval.FuncDef, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.funcs[name]
	return fn, ok
}

// Has checks if a function exists.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.funcs[name]
	return ok
}

// List returns all registered function names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.funcs))
	for name := range r.funcs {
		names = append(names, name)
	}
	return names
}

// Global registry functions

// Register adds a function to the global registry.
func Register(def eval.FuncDef) {
	globalRegistry.Register(def)
}

// Get retrieves a function from the global registry.
func Get(name string) (eval.FuncDef, bool) {
	return globalRegistry.Get(name)
}

// Has checks if a function exists in the global registry.
func Has(name string) bool {
	return globalRegistry.Has(name)
}

// List returns all function names from the global registry.
func List() []string {
	return globalRegistry.List()
}

// GetRegistry returns the global registry.
func GetRegistry() *Registry {
	return globalRegistry
}
