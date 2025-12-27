package eval

import (
	"context"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/parser/grammar"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// FuncImpl is the signature for function implementations.
type FuncImpl func(ctx *Context, input types.Collection, args []interface{}) (types.Collection, error)

// FuncDef defines a FHIRPath function.
type FuncDef struct {
	Name    string
	MinArgs int
	MaxArgs int
	Fn      FuncImpl
}

// FuncRegistry is an interface for function lookup.
type FuncRegistry interface {
	Get(name string) (FuncDef, bool)
}

// Resolver handles FHIR reference resolution.
type Resolver interface {
	Resolve(ctx context.Context, reference string) ([]byte, error)
}

// Evaluator evaluates FHIRPath expressions using the visitor pattern.
type Evaluator struct {
	grammar.BasefhirpathVisitor
	ctx   *Context
	funcs FuncRegistry
}

// Context holds the evaluation state.
type Context struct {
	root      types.Collection
	this      types.Collection
	index     int
	total     types.Value
	variables map[string]types.Collection
	limits    map[string]int
	goCtx     context.Context
	resolver  Resolver
}

// NewContext creates a new evaluation context.
// Automatically sets %resource and %context to the root resource for FHIR constraint evaluation.
// Per FHIRPath spec:
//   - %resource: the root resource being evaluated
//   - %context: the original node passed to the evaluation engine (same as %resource for top-level evaluation)
func NewContext(resource []byte) *Context {
	//nolint:errcheck // Empty collection is acceptable for invalid JSON in context creation
	root, _ := types.JSONToCollection(resource)

	// Initialize variables map with %resource and %context pointing to root
	// %resource is required by FHIR constraints like bdl-3, bdl-4
	// %context represents the evaluation context (same as root for top-level evaluation)
	variables := make(map[string]types.Collection)
	variables["resource"] = root
	variables["context"] = root

	return &Context{
		root:      root,
		this:      root,
		variables: variables,
		limits:    make(map[string]int),
		goCtx:     context.Background(),
	}
}

// SetLimit sets a limit value (e.g., maxDepth, maxCollectionSize).
func (c *Context) SetLimit(name string, value int) {
	if c.limits == nil {
		c.limits = make(map[string]int)
	}
	c.limits[name] = value
}

// GetLimit gets a limit value.
func (c *Context) GetLimit(name string) int {
	if c.limits == nil {
		return 0
	}
	return c.limits[name]
}

// SetContext sets the Go context for cancellation.
func (c *Context) SetContext(ctx context.Context) {
	c.goCtx = ctx
}

// Context returns the Go context.
func (c *Context) Context() context.Context {
	if c.goCtx == nil {
		return context.Background()
	}
	return c.goCtx
}

// SetResolver sets the reference resolver.
func (c *Context) SetResolver(r Resolver) {
	c.resolver = r
}

// GetResolver returns the reference resolver.
func (c *Context) GetResolver() Resolver {
	return c.resolver
}

// CheckCancellation checks if the context has been canceled.
func (c *Context) CheckCancellation() error {
	if c.goCtx == nil {
		return nil
	}
	select {
	case <-c.goCtx.Done():
		return c.goCtx.Err()
	default:
		return nil
	}
}

// CheckCollectionSize validates that a collection doesn't exceed the maximum size.
// Returns an error if the collection is too large.
func (c *Context) CheckCollectionSize(col types.Collection) error {
	maxSize := c.GetLimit("maxCollectionSize")
	if maxSize > 0 && len(col) > maxSize {
		return NewEvalError(ErrInvalidExpression,
			"collection size %d exceeds maximum allowed %d", len(col), maxSize)
	}
	return nil
}

// EnforceCollectionLimit truncates a collection if it exceeds the maximum size.
// Returns the (possibly truncated) collection and whether truncation occurred.
func (c *Context) EnforceCollectionLimit(col types.Collection) (types.Collection, bool) {
	maxSize := c.GetLimit("maxCollectionSize")
	if maxSize > 0 && len(col) > maxSize {
		return col[:maxSize], true
	}
	return col, false
}

// Root returns the root collection.
func (c *Context) Root() types.Collection {
	return c.root
}

// This returns the current $this value.
func (c *Context) This() types.Collection {
	return c.this
}

// WithThis returns a new context with the given $this value.
func (c *Context) WithThis(this types.Collection) *Context {
	newCtx := *c
	newCtx.this = this
	return &newCtx
}

// WithIndex returns a new context with the given $index value.
func (c *Context) WithIndex(index int) *Context {
	newCtx := *c
	newCtx.index = index
	return &newCtx
}

// SetVariable sets an external variable.
func (c *Context) SetVariable(name string, value types.Collection) {
	c.variables[name] = value
}

// GetVariable gets an external variable.
func (c *Context) GetVariable(name string) (types.Collection, bool) {
	v, ok := c.variables[name]
	return v, ok
}

// NewEvaluator creates a new evaluator with the given context and function registry.
func NewEvaluator(ctx *Context, funcs FuncRegistry) *Evaluator {
	return &Evaluator{ctx: ctx, funcs: funcs}
}

// Evaluate evaluates a parse tree and returns the result.
func (e *Evaluator) Evaluate(tree antlr.ParseTree) (types.Collection, error) {
	result := e.Visit(tree)
	if err, ok := result.(error); ok {
		return nil, err
	}
	if col, ok := result.(types.Collection); ok {
		return col, nil
	}
	return types.Collection{}, nil
}

// Visit dispatches to the appropriate visitor method.
func (e *Evaluator) Visit(tree antlr.ParseTree) interface{} {
	if tree == nil {
		return types.Collection{}
	}
	return tree.Accept(e)
}

// VisitEntireExpression visits the root expression.
func (e *Evaluator) VisitEntireExpression(ctx *grammar.EntireExpressionContext) interface{} {
	return e.Visit(ctx.Expression())
}

// VisitTermExpression visits a term expression.
func (e *Evaluator) VisitTermExpression(ctx *grammar.TermExpressionContext) interface{} {
	return e.Visit(ctx.Term())
}

// VisitInvocationTerm visits an invocation term.
func (e *Evaluator) VisitInvocationTerm(ctx *grammar.InvocationTermContext) interface{} {
	return e.Visit(ctx.Invocation())
}

// VisitLiteralTerm visits a literal term.
func (e *Evaluator) VisitLiteralTerm(ctx *grammar.LiteralTermContext) interface{} {
	return e.Visit(ctx.Literal())
}

// VisitParenthesizedTerm visits a parenthesized expression.
func (e *Evaluator) VisitParenthesizedTerm(ctx *grammar.ParenthesizedTermContext) interface{} {
	return e.Visit(ctx.Expression())
}

// VisitExternalConstantTerm visits an external constant.
func (e *Evaluator) VisitExternalConstantTerm(ctx *grammar.ExternalConstantTermContext) interface{} {
	return e.Visit(ctx.ExternalConstant())
}

// VisitExternalConstant visits an external constant (%name).
func (e *Evaluator) VisitExternalConstant(ctx *grammar.ExternalConstantContext) interface{} {
	var name string
	if ctx.Identifier() != nil {
		name = stripBackticks(ctx.Identifier().GetText())
	} else if ctx.STRING() != nil {
		name = unquoteString(ctx.STRING().GetText())
	}

	if value, ok := e.ctx.GetVariable(name); ok {
		return value
	}
	return NewEvalError(ErrInvalidPath, "undefined variable: %"+name)
}

// Literal visitors

// VisitNullLiteral visits a null literal {}.
func (e *Evaluator) VisitNullLiteral(ctx *grammar.NullLiteralContext) interface{} {
	return types.Collection{}
}

// VisitBooleanLiteral visits a boolean literal.
func (e *Evaluator) VisitBooleanLiteral(ctx *grammar.BooleanLiteralContext) interface{} {
	text := ctx.GetText()
	return types.Collection{types.NewBoolean(text == "true")}
}

// VisitStringLiteral visits a string literal.
func (e *Evaluator) VisitStringLiteral(ctx *grammar.StringLiteralContext) interface{} {
	text := unquoteString(ctx.STRING().GetText())
	return types.Collection{types.NewString(text)}
}

// VisitNumberLiteral visits a number literal.
func (e *Evaluator) VisitNumberLiteral(ctx *grammar.NumberLiteralContext) interface{} {
	text := ctx.NUMBER().GetText()

	// Check if it's an integer
	if !strings.Contains(text, ".") {
		if i, err := strconv.ParseInt(text, 10, 64); err == nil {
			return types.Collection{types.NewInteger(i)}
		}
	}

	// Parse as decimal
	d, err := types.NewDecimal(text)
	if err != nil {
		return ParseError("invalid number: " + text)
	}
	return types.Collection{d}
}

// VisitDateLiteral visits a date literal.
func (e *Evaluator) VisitDateLiteral(ctx *grammar.DateLiteralContext) interface{} {
	text := ctx.DATE().GetText()
	// Remove the @ prefix
	if text != "" && text[0] == '@' {
		text = text[1:]
	}
	d, err := types.NewDate(text)
	if err != nil {
		return ParseError("invalid date: " + text)
	}
	return types.Collection{d}
}

// VisitDateTimeLiteral visits a datetime literal.
func (e *Evaluator) VisitDateTimeLiteral(ctx *grammar.DateTimeLiteralContext) interface{} {
	text := ctx.DATETIME().GetText()
	// Remove the @ prefix
	if text != "" && text[0] == '@' {
		text = text[1:]
	}
	dt, err := types.NewDateTime(text)
	if err != nil {
		return ParseError("invalid datetime: " + text)
	}
	return types.Collection{dt}
}

// VisitTimeLiteral visits a time literal.
func (e *Evaluator) VisitTimeLiteral(ctx *grammar.TimeLiteralContext) interface{} {
	text := ctx.TIME().GetText()
	// Remove the @ prefix
	if text != "" && text[0] == '@' {
		text = text[1:]
	}
	t, err := types.NewTime(text)
	if err != nil {
		return ParseError("invalid time: " + text)
	}
	return types.Collection{t}
}

// VisitQuantityLiteral visits a quantity literal.
func (e *Evaluator) VisitQuantityLiteral(ctx *grammar.QuantityLiteralContext) interface{} {
	text := ctx.GetText()
	q, err := types.NewQuantity(text)
	if err != nil {
		return ParseError("invalid quantity: " + text)
	}
	return types.Collection{q}
}

// Invocation visitors

// VisitMemberInvocation visits a member access.
func (e *Evaluator) VisitMemberInvocation(ctx *grammar.MemberInvocationContext) interface{} {
	name := stripBackticks(ctx.Identifier().GetText())
	return e.navigateMember(e.ctx.This(), name)
}

// VisitFunctionInvocation visits a function call.
func (e *Evaluator) VisitFunctionInvocation(ctx *grammar.FunctionInvocationContext) interface{} {
	funcCtx := ctx.Function()
	name := stripBackticks(funcCtx.Identifier().GetText())

	// Get function from registry
	fn, ok := e.funcs.Get(name)
	if !ok {
		return FunctionNotFoundError(name)
	}

	// Validate argument count
	paramList := funcCtx.ParamList()
	argCount := 0
	var argExprs []grammar.IExpressionContext
	if paramList != nil {
		argExprs = paramList.AllExpression()
		argCount = len(argExprs)
	}

	if argCount < fn.MinArgs {
		return InvalidArgumentsError(name, fn.MinArgs, argCount)
	}
	if fn.MaxArgs >= 0 && argCount > fn.MaxArgs {
		return InvalidArgumentsError(name, fn.MaxArgs, argCount)
	}

	// Handle special functions that need per-element evaluation
	input := e.ctx.This()
	switch name {
	case "where":
		if argCount > 0 {
			return e.evaluateWhere(input, argExprs[0])
		}
	case "exists":
		if argCount > 0 {
			return e.evaluateExists(input, argExprs[0])
		}
	case "all":
		if argCount > 0 {
			return e.evaluateAll(input, argExprs[0])
		}
	case "select":
		if argCount > 0 {
			return e.evaluateSelect(input, argExprs[0])
		}
	case "is":
		if argCount > 0 {
			return e.evaluateIsFunction(input, argExprs[0])
		}
	case "as":
		if argCount > 0 {
			return e.evaluateAsFunction(input, argExprs[0])
		}
	case "ofType":
		if argCount > 0 {
			return e.evaluateOfType(input, argExprs[0])
		}
	case "iif":
		// iif requires lazy evaluation - only evaluate the branch that matches
		if argCount >= 2 {
			return e.evaluateIif(input, argExprs)
		}
	}

	// Evaluate arguments normally
	args := make([]interface{}, argCount)
	for i, argExpr := range argExprs {
		result := e.Visit(argExpr)
		if err, ok := result.(error); ok {
			return err
		}
		args[i] = result
	}

	// Call the function
	result, err := fn.Fn(e.ctx, e.ctx.This(), args)
	if err != nil {
		return err
	}
	return result
}

// evaluateWhere evaluates the where() function with per-element criteria.
func (e *Evaluator) evaluateWhere(input types.Collection, criteria grammar.IExpressionContext) interface{} {
	result := types.Collection{}

	// Check collection size limit
	if err := e.ctx.CheckCollectionSize(input); err != nil {
		return err
	}

	for i, item := range input {
		// Check for cancellation periodically (every 100 iterations)
		if i%100 == 0 {
			if err := e.ctx.CheckCancellation(); err != nil {
				return err
			}
		}

		// Set $this to current item and $index
		oldThis := e.ctx.this
		oldIndex := e.ctx.index
		e.ctx.this = types.Collection{item}
		e.ctx.index = i

		// Evaluate the criteria
		criteriaResult := e.Visit(criteria)

		// Restore context
		e.ctx.this = oldThis
		e.ctx.index = oldIndex

		if err, ok := criteriaResult.(error); ok {
			return err
		}

		// Check if criteria is true
		if col, ok := criteriaResult.(types.Collection); ok && !col.Empty() {
			if b, ok := col[0].(types.Boolean); ok && b.Bool() {
				result = append(result, item)
			}
		}
	}

	return result
}

// evaluateExists evaluates exists() with optional criteria.
func (e *Evaluator) evaluateExists(input types.Collection, criteria grammar.IExpressionContext) interface{} {
	for i, item := range input {
		// Check for cancellation periodically
		if i%100 == 0 {
			if err := e.ctx.CheckCancellation(); err != nil {
				return err
			}
		}

		// Set $this to current item
		oldThis := e.ctx.this
		oldIndex := e.ctx.index
		e.ctx.this = types.Collection{item}
		e.ctx.index = i

		// Evaluate the criteria
		criteriaResult := e.Visit(criteria)

		// Restore context
		e.ctx.this = oldThis
		e.ctx.index = oldIndex

		if err, ok := criteriaResult.(error); ok {
			return err
		}

		// Check if criteria is true
		if col, ok := criteriaResult.(types.Collection); ok && !col.Empty() {
			if b, ok := col[0].(types.Boolean); ok && b.Bool() {
				return types.Collection{types.NewBoolean(true)}
			}
		}
	}

	return types.Collection{types.NewBoolean(false)}
}

// evaluateAll evaluates all() - returns true if all elements match criteria.
func (e *Evaluator) evaluateAll(input types.Collection, criteria grammar.IExpressionContext) interface{} {
	if input.Empty() {
		return types.Collection{types.NewBoolean(true)}
	}

	for i, item := range input {
		// Check for cancellation periodically
		if i%100 == 0 {
			if err := e.ctx.CheckCancellation(); err != nil {
				return err
			}
		}

		// Set $this to current item
		oldThis := e.ctx.this
		oldIndex := e.ctx.index
		e.ctx.this = types.Collection{item}
		e.ctx.index = i

		// Evaluate the criteria
		criteriaResult := e.Visit(criteria)

		// Restore context
		e.ctx.this = oldThis
		e.ctx.index = oldIndex

		if err, ok := criteriaResult.(error); ok {
			return err
		}

		// Check if criteria is true
		if col, ok := criteriaResult.(types.Collection); ok {
			if col.Empty() {
				return types.Collection{types.NewBoolean(false)}
			}
			if b, ok := col[0].(types.Boolean); ok && !b.Bool() {
				return types.Collection{types.NewBoolean(false)}
			}
		}
	}

	return types.Collection{types.NewBoolean(true)}
}

// evaluateSelect evaluates select() - projects each element.
func (e *Evaluator) evaluateSelect(input types.Collection, projection grammar.IExpressionContext) interface{} {
	result := types.Collection{}

	// Check collection size limit
	if err := e.ctx.CheckCollectionSize(input); err != nil {
		return err
	}

	for i, item := range input {
		// Check for cancellation periodically
		if i%100 == 0 {
			if err := e.ctx.CheckCancellation(); err != nil {
				return err
			}
		}

		// Set $this to current item
		oldThis := e.ctx.this
		oldIndex := e.ctx.index
		e.ctx.this = types.Collection{item}
		e.ctx.index = i

		// Evaluate the projection
		projResult := e.Visit(projection)

		// Restore context
		e.ctx.this = oldThis
		e.ctx.index = oldIndex

		if err, ok := projResult.(error); ok {
			return err
		}

		// Add projection result to output
		if col, ok := projResult.(types.Collection); ok {
			result = append(result, col...)

			// Check if result is getting too large
			if err := e.ctx.CheckCollectionSize(result); err != nil {
				return err
			}
		}
	}

	return result
}

// evaluateIsFunction evaluates is() function - checks if input is of specified type.
// This handles is(Type) where Type is an identifier like Composition, Patient, etc.
func (e *Evaluator) evaluateIsFunction(input types.Collection, typeExpr grammar.IExpressionContext) interface{} {
	// Empty input returns empty
	if input.Empty() {
		return types.Collection{}
	}

	// is() requires singleton input
	if len(input) != 1 {
		return SingletonError(len(input))
	}

	// Extract the type name from the expression
	typeName := e.extractTypeNameFromExpr(typeExpr)
	if typeName == "" {
		return InvalidArgumentsError("is", 1, 0)
	}

	// Get actual type - Type() already returns resourceType for ObjectValue
	actualType := input[0].Type()

	matches := TypeMatches(actualType, typeName)
	return types.Collection{types.NewBoolean(matches)}
}

// evaluateAsFunction evaluates as() function - casts input to specified type.
// Returns input if it matches the type, empty otherwise.
func (e *Evaluator) evaluateAsFunction(input types.Collection, typeExpr grammar.IExpressionContext) interface{} {
	// Empty input returns empty
	if input.Empty() {
		return types.Collection{}
	}

	// as() requires singleton input
	if len(input) != 1 {
		return SingletonError(len(input))
	}

	// Extract the type name from the expression
	typeName := e.extractTypeNameFromExpr(typeExpr)
	if typeName == "" {
		return InvalidArgumentsError("as", 1, 0)
	}

	// Get actual type - Type() already returns resourceType for ObjectValue
	actualType := input[0].Type()

	if TypeMatches(actualType, typeName) {
		return input
	}

	return types.Collection{}
}

// extractTypeNameFromExpr extracts a type name from a FHIRPath expression.
// Handles identifiers like Composition, Patient, and qualified names like FHIR.Patient.
func (e *Evaluator) extractTypeNameFromExpr(expr grammar.IExpressionContext) string {
	// Get the text of the expression directly - this handles simple identifiers
	text := expr.GetText()
	if text != "" {
		return text
	}
	return ""
}

// evaluateOfType evaluates ofType() function - filters collection by type.
// Unlike is()/as() which require singleton, ofType() works on collections.
func (e *Evaluator) evaluateOfType(input types.Collection, typeExpr grammar.IExpressionContext) interface{} {
	// Empty input returns empty
	if input.Empty() {
		return types.Collection{}
	}

	// Extract the type name from the expression
	typeName := e.extractTypeNameFromExpr(typeExpr)
	if typeName == "" {
		return InvalidArgumentsError("ofType", 1, 0)
	}

	result := types.Collection{}
	for _, item := range input {
		actualType := item.Type()

		// For ObjectValue, also check if it's a FHIR type matching the request
		if obj, ok := item.(*types.ObjectValue); ok {
			// Try to get more specific type information
			actualType = obj.Type()
		}

		if TypeMatches(actualType, typeName) {
			result = append(result, item)
		}
	}

	return result
}

// evaluateIif evaluates the iif() function with lazy evaluation.
// Only the matching branch is evaluated, preventing errors from the other branch.
// Signature: iif(criterion, true-result [, otherwise-result])
func (e *Evaluator) evaluateIif(_ types.Collection, argExprs []grammar.IExpressionContext) interface{} {
	if len(argExprs) < 2 {
		return InvalidArgumentsError("iif", 2, len(argExprs))
	}

	// Evaluate the criterion (first argument)
	criterionResult := e.Visit(argExprs[0])
	if err, ok := criterionResult.(error); ok {
		return err
	}

	// Convert criterion to boolean
	criterion := false
	if coll, ok := criterionResult.(types.Collection); ok {
		if !coll.Empty() {
			if b, ok := coll[0].(types.Boolean); ok {
				criterion = b.Bool()
			}
		}
	}

	// Lazily evaluate only the matching branch
	if criterion {
		// Evaluate and return true-result (second argument)
		result := e.Visit(argExprs[1])
		if err, ok := result.(error); ok {
			return err
		}
		if coll, ok := result.(types.Collection); ok {
			return coll
		}
		return types.Collection{}
	}

	// Evaluate and return otherwise-result (third argument) if provided
	if len(argExprs) > 2 {
		result := e.Visit(argExprs[2])
		if err, ok := result.(error); ok {
			return err
		}
		if coll, ok := result.(types.Collection); ok {
			return coll
		}
	}

	return types.Collection{}
}

// VisitThisInvocation visits $this.
func (e *Evaluator) VisitThisInvocation(ctx *grammar.ThisInvocationContext) interface{} {
	return e.ctx.This()
}

// VisitIndexInvocation visits $index.
func (e *Evaluator) VisitIndexInvocation(ctx *grammar.IndexInvocationContext) interface{} {
	return types.Collection{types.NewInteger(int64(e.ctx.index))}
}

// VisitTotalInvocation visits $total.
func (e *Evaluator) VisitTotalInvocation(ctx *grammar.TotalInvocationContext) interface{} {
	if e.ctx.total != nil {
		return types.Collection{e.ctx.total}
	}
	return types.Collection{}
}

// Expression visitors

// VisitInvocationExpression visits expr.invocation.
func (e *Evaluator) VisitInvocationExpression(ctx *grammar.InvocationExpressionContext) interface{} {
	// Evaluate the base expression
	base := e.Visit(ctx.Expression())
	if err, ok := base.(error); ok {
		return err
	}
	baseCol := base.(types.Collection)

	// Save current this and set new this
	oldThis := e.ctx.this
	e.ctx.this = baseCol
	defer func() { e.ctx.this = oldThis }()

	// Evaluate the invocation
	return e.Visit(ctx.Invocation())
}

// VisitIndexerExpression visits expr[index].
func (e *Evaluator) VisitIndexerExpression(ctx *grammar.IndexerExpressionContext) interface{} {
	base := e.Visit(ctx.Expression(0))
	if err, ok := base.(error); ok {
		return err
	}
	baseCol := base.(types.Collection)

	index := e.Visit(ctx.Expression(1))
	if err, ok := index.(error); ok {
		return err
	}
	indexCol := index.(types.Collection)

	if indexCol.Empty() {
		return types.Collection{}
	}

	// Get index as integer
	idx, ok := indexCol[0].(types.Integer)
	if !ok {
		return TypeError("Integer", indexCol[0].Type(), "indexer")
	}

	i := int(idx.Value())
	if i < 0 || i >= len(baseCol) {
		return types.Collection{}
	}

	return types.Collection{baseCol[i]}
}

// VisitPolarityExpression visits +expr or -expr.
func (e *Evaluator) VisitPolarityExpression(ctx *grammar.PolarityExpressionContext) interface{} {
	result := e.Visit(ctx.Expression())
	if err, ok := result.(error); ok {
		return err
	}
	col := result.(types.Collection)

	if col.Empty() {
		return col
	}
	if len(col) != 1 {
		return SingletonError(len(col))
	}

	// Check if it's negation
	if ctx.GetChild(0).(antlr.TerminalNode).GetText() == "-" {
		negated, err := Negate(col[0])
		if err != nil {
			return err
		}
		return types.Collection{negated}
	}

	return col
}

// VisitMultiplicativeExpression visits expr * expr, expr / expr, etc.
func (e *Evaluator) VisitMultiplicativeExpression(ctx *grammar.MultiplicativeExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	// Empty propagation
	if leftCol.Empty() || rightCol.Empty() {
		return types.Collection{}
	}

	// Singleton check
	if len(leftCol) != 1 || len(rightCol) != 1 {
		return SingletonError(len(leftCol) + len(rightCol))
	}

	op := ctx.GetChild(1).(antlr.TerminalNode).GetText()

	var result types.Value
	var err error

	switch op {
	case "*":
		result, err = Multiply(leftCol[0], rightCol[0])
	case "/":
		result, err = Divide(leftCol[0], rightCol[0])
	case "div":
		result, err = IntegerDivide(leftCol[0], rightCol[0])
	case "mod":
		result, err = Modulo(leftCol[0], rightCol[0])
	}

	if err != nil {
		return err
	}
	return types.Collection{result}
}

// VisitAdditiveExpression visits expr + expr, expr - expr, expr & expr.
func (e *Evaluator) VisitAdditiveExpression(ctx *grammar.AdditiveExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	op := ctx.GetChild(1).(antlr.TerminalNode).GetText()

	// String concatenation with & handles empty as empty string
	if op == "&" {
		return Concatenate(leftCol, rightCol)
	}

	// Empty propagation for + and -
	if leftCol.Empty() || rightCol.Empty() {
		return types.Collection{}
	}

	// Singleton check
	if len(leftCol) != 1 || len(rightCol) != 1 {
		return SingletonError(len(leftCol) + len(rightCol))
	}

	var result types.Value
	var err error

	switch op {
	case "+":
		result, err = Add(leftCol[0], rightCol[0])
	case "-":
		result, err = Subtract(leftCol[0], rightCol[0])
	}

	if err != nil {
		return err
	}
	return types.Collection{result}
}

// VisitUnionExpression visits expr | expr.
func (e *Evaluator) VisitUnionExpression(ctx *grammar.UnionExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	return Union(leftCol, rightCol)
}

// VisitInequalityExpression visits comparison expressions.
func (e *Evaluator) VisitInequalityExpression(ctx *grammar.InequalityExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	// Empty propagation
	if leftCol.Empty() || rightCol.Empty() {
		return types.Collection{}
	}

	// Singleton check
	if len(leftCol) != 1 || len(rightCol) != 1 {
		return SingletonError(len(leftCol) + len(rightCol))
	}

	op := ctx.GetChild(1).(antlr.TerminalNode).GetText()

	var result types.Collection
	var err error

	switch op {
	case "<":
		result, err = LessThan(leftCol[0], rightCol[0])
	case "<=":
		result, err = LessOrEqual(leftCol[0], rightCol[0])
	case ">":
		result, err = GreaterThan(leftCol[0], rightCol[0])
	case ">=":
		result, err = GreaterOrEqual(leftCol[0], rightCol[0])
	default:
		return types.Collection{}
	}

	if err != nil {
		return err
	}
	return result
}

// VisitEqualityExpression visits equality expressions.
func (e *Evaluator) VisitEqualityExpression(ctx *grammar.EqualityExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	op := ctx.GetChild(1).(antlr.TerminalNode).GetText()

	switch op {
	case "=":
		return Equal(leftCol, rightCol)
	case "!=":
		return NotEqual(leftCol, rightCol)
	case "~":
		return Equivalent(leftCol, rightCol)
	case "!~":
		return NotEquivalent(leftCol, rightCol)
	}

	return types.Collection{}
}

// VisitMembershipExpression visits 'in' and 'contains' expressions.
func (e *Evaluator) VisitMembershipExpression(ctx *grammar.MembershipExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	op := ctx.GetChild(1).(antlr.TerminalNode).GetText()

	switch op {
	case "in":
		return In(leftCol, rightCol)
	case "contains":
		return Contains(leftCol, rightCol)
	}

	return types.Collection{}
}

// VisitAndExpression visits expr and expr.
func (e *Evaluator) VisitAndExpression(ctx *grammar.AndExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	return And(leftCol, rightCol)
}

// VisitOrExpression visits expr or expr, expr xor expr.
func (e *Evaluator) VisitOrExpression(ctx *grammar.OrExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	op := ctx.GetChild(1).(antlr.TerminalNode).GetText()

	switch op {
	case "or":
		return Or(leftCol, rightCol)
	case "xor":
		return Xor(leftCol, rightCol)
	}

	return types.Collection{}
}

// VisitImpliesExpression visits expr implies expr.
func (e *Evaluator) VisitImpliesExpression(ctx *grammar.ImpliesExpressionContext) interface{} {
	left := e.Visit(ctx.Expression(0))
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	right := e.Visit(ctx.Expression(1))
	if err, ok := right.(error); ok {
		return err
	}
	rightCol := right.(types.Collection)

	return Implies(leftCol, rightCol)
}

// VisitTypeExpression visits 'is' and 'as' expressions.
func (e *Evaluator) VisitTypeExpression(ctx *grammar.TypeExpressionContext) interface{} {
	left := e.Visit(ctx.Expression())
	if err, ok := left.(error); ok {
		return err
	}
	leftCol := left.(types.Collection)

	typeName := ctx.TypeSpecifier().GetText()
	op := ctx.GetChild(1).(antlr.TerminalNode).GetText()

	if leftCol.Empty() {
		return types.Collection{}
	}

	if len(leftCol) != 1 {
		return SingletonError(len(leftCol))
	}

	actualType := leftCol[0].Type()

	switch op {
	case "is":
		return types.Collection{types.NewBoolean(TypeMatches(actualType, typeName))}
	case "as":
		if TypeMatches(actualType, typeName) {
			return leftCol
		}
		return types.Collection{}
	}

	return types.Collection{}
}

// nonDomainResources contains FHIR resources that inherit directly from Resource,
// not from DomainResource. All other resources inherit from DomainResource.
var nonDomainResources = map[string]bool{
	"Bundle":     true,
	"Binary":     true,
	"Parameters": true,
}

// IsDomainResource returns true if the given resource type inherits from DomainResource.
// Bundle, Binary, and Parameters inherit directly from Resource, not DomainResource.
func IsDomainResource(resourceType string) bool {
	return !nonDomainResources[resourceType]
}

// IsSubtypeOf checks if actualType is a subtype of (or equal to) baseType.
// This handles the FHIR type hierarchy:
//
//	Resource
//	  └── DomainResource
//	        ├── Patient
//	        ├── Observation
//	        └── ... (most resources)
//	  └── Bundle, Binary, Parameters (directly inherit from Resource)
func IsSubtypeOf(actualType, baseType string) bool {
	// Direct match
	if actualType == baseType {
		return true
	}

	// Case-insensitive direct match
	if strings.EqualFold(actualType, baseType) {
		return true
	}

	// Check Resource base type - all resources inherit from Resource
	if baseType == "Resource" || strings.EqualFold(baseType, "resource") {
		// Any non-empty type that looks like a resource type matches Resource
		// Resource types are PascalCase and don't include primitives
		return isPossibleResourceType(actualType)
	}

	// Check DomainResource base type
	if baseType == "DomainResource" || strings.EqualFold(baseType, "domainresource") {
		// Most resources inherit from DomainResource, except Bundle, Binary, Parameters
		return isPossibleResourceType(actualType) && IsDomainResource(actualType)
	}

	return false
}

// isPossibleResourceType checks if the type looks like a FHIR resource type.
// Resource types are PascalCase and are not primitive types.
func isPossibleResourceType(typeName string) bool {
	if typeName == "" {
		return false
	}

	// Primitive types are not resources
	primitiveTypes := map[string]bool{
		"Boolean": true, "String": true, "Integer": true, "Decimal": true,
		"Date": true, "DateTime": true, "Time": true, "Quantity": true,
		"Object": true,
	}
	if primitiveTypes[typeName] {
		return false
	}

	// Resource types start with uppercase
	return typeName[0] >= 'A' && typeName[0] <= 'Z'
}

// TypeMatches checks if actualType matches the requested typeName.
// Handles case-insensitive comparison and FHIR type aliases.
// This function is exported for use by the is() function implementation.
func TypeMatches(actualType, typeName string) bool {
	// Direct match
	if actualType == typeName {
		return true
	}

	// Normalize to lowercase for comparison
	actualLower := strings.ToLower(actualType)
	typeNameLower := strings.ToLower(typeName)

	// Case-insensitive match
	if actualLower == typeNameLower {
		return true
	}

	// Check FHIR base type inheritance (Resource, DomainResource)
	if IsSubtypeOf(actualType, typeName) {
		return true
	}

	// FHIR primitive type mappings (FHIR uses lowercase, FHIRPath uses PascalCase)
	fhirToFHIRPath := map[string]string{
		"boolean":        "Boolean",
		"string":         "String",
		"integer":        "Integer",
		"decimal":        "Decimal",
		"date":           "Date",
		"datetime":       "DateTime",
		"time":           "Time",
		"instant":        "DateTime",
		"uri":            "String",
		"url":            "String",
		"canonical":      "String",
		"base64binary":   "String",
		"code":           "String",
		"id":             "String",
		"markdown":       "String",
		"oid":            "String",
		"uuid":           "String",
		"positiveint":    "Integer",
		"unsignedint":    "Integer",
		"integer64":      "Integer",
		"quantity":       "Quantity",
		"simplequantity": "Quantity",
		"age":            "Quantity",
		"count":          "Quantity",
		"distance":       "Quantity",
		"duration":       "Quantity",
		"money":          "Quantity",
	}

	// Check if requesting a FHIR type that maps to a FHIRPath type
	if fhirPathType, ok := fhirToFHIRPath[typeNameLower]; ok {
		if actualType == fhirPathType {
			return true
		}
	}

	// Check reverse: if actual type is a FHIR type that maps to the requested FHIRPath type
	if fhirPathType, ok := fhirToFHIRPath[actualLower]; ok {
		if fhirPathType == typeName || strings.EqualFold(fhirPathType, typeName) {
			return true
		}
	}

	// System type namespace handling (FHIR.* and System.*)
	// System.Boolean, System.String, etc.
	if strings.HasPrefix(typeNameLower, "system.") {
		systemType := typeName[7:] // Remove "System." prefix
		if strings.EqualFold(actualType, systemType) {
			return true
		}
	}

	// FHIR namespace handling
	if strings.HasPrefix(typeNameLower, "fhir.") {
		fhirType := typeName[5:] // Remove "FHIR." prefix
		if strings.EqualFold(actualType, fhirType) {
			return true
		}
	}

	return false
}

// Helper functions

// polymorphicTypeSuffixes contains all FHIR type suffixes for polymorphic elements (value[x] pattern).
// These are used to resolve element names like "value" to "valueQuantity", "valueString", etc.
var polymorphicTypeSuffixes = []string{
	// Primitive types
	"Boolean", "Integer", "Integer64", "Decimal", "String", "Code", "Id", "Uri", "Url", "Canonical",
	"Base64Binary", "Instant", "Date", "DateTime", "Time", "Oid", "Uuid", "Markdown", "PositiveInt", "UnsignedInt",
	// Complex types
	"Quantity", "CodeableConcept", "Coding", "Range", "Period", "Ratio", "RatioRange",
	"Identifier", "Reference", "Attachment", "HumanName", "Address", "ContactPoint",
	"Timing", "Signature", "Annotation", "SampledData", "Age", "Distance", "Duration",
	"Count", "Money", "MoneyQuantity", "SimpleQuantity",
	// Special types
	"Meta", "Dosage", "ContactDetail", "Contributor", "DataRequirement", "Expression",
	"ParameterDefinition", "RelatedArtifact", "TriggerDefinition", "UsageContext",
}

// navigateMember navigates to a member of objects in the collection.
// Supports FHIR polymorphic elements (value[x] pattern) by automatically
// resolving element names like "value" to their typed variants.
func (e *Evaluator) navigateMember(input types.Collection, name string) types.Collection {
	result := types.Collection{}

	for _, item := range input {
		obj, ok := item.(*types.ObjectValue)
		if !ok {
			continue
		}

		// Check if name matches resourceType (for FHIR resources)
		// Uses IsSubtypeOf to handle Resource and DomainResource base types
		if IsSubtypeOf(obj.Type(), name) {
			result = append(result, obj)
			continue
		}

		// Try direct field access first
		children := obj.GetCollection(name)
		if len(children) > 0 {
			result = append(result, children...)
			continue
		}

		// If direct access failed, try polymorphic element resolution
		// This handles FHIR's value[x] pattern where "value" can resolve to
		// "valueQuantity", "valueString", "valueCodeableConcept", etc.
		polymorphicChildren := e.resolvePolymorphicField(obj, name)
		result = append(result, polymorphicChildren...)
	}

	return result
}

// resolvePolymorphicField attempts to resolve a polymorphic FHIR element.
// For example, accessing "value" will search for "valueQuantity", "valueString", etc.
func (e *Evaluator) resolvePolymorphicField(obj *types.ObjectValue, name string) types.Collection {
	result := types.Collection{}

	// Try each possible type suffix
	for _, suffix := range polymorphicTypeSuffixes {
		fieldName := name + suffix
		children := obj.GetCollection(fieldName)
		if len(children) > 0 {
			result = append(result, children...)
			// Return on first match - polymorphic elements have only one variant
			return result
		}
	}

	return result
}

// unquoteString removes quotes and handles escape sequences.
func unquoteString(s string) string {
	if len(s) < 2 {
		return s
	}
	// Remove surrounding quotes
	s = s[1 : len(s)-1]

	// Handle escape sequences
	s = strings.ReplaceAll(s, "\\'", "'")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\t", "\t")

	return s
}

// stripBackticks removes backtick delimiters from delimited identifiers.
// FHIRPath allows backticks for identifiers with special characters: `PID-1`
func stripBackticks(s string) string {
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		return s[1 : len(s)-1]
	}
	return s
}
