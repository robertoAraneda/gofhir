package fhirpath

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/parser/grammar"
)

// errorListener captures parsing errors.
type errorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

func (l *errorListener) SyntaxError(_ antlr.Recognizer, _ interface{}, line, column int, msg string, _ antlr.RecognitionException) {
	l.errors = append(l.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

// compile parses a FHIRPath expression into a compiled Expression.
func compile(expr string) (*Expression, error) {
	if expr == "" {
		return nil, fmt.Errorf("empty expression")
	}

	// Create lexer
	input := antlr.NewInputStream(expr)
	lexer := grammar.NewfhirpathLexer(input)

	// Set up error listener for lexer
	lexerErrors := &errorListener{}
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(lexerErrors)

	// Create parser
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := grammar.NewfhirpathParser(stream)

	// Set up error listener for parser
	parserErrors := &errorListener{}
	parser.RemoveErrorListeners()
	parser.AddErrorListener(parserErrors)

	// Parse the expression
	tree := parser.EntireExpression()

	// Check for errors
	if len(lexerErrors.errors) > 0 {
		return nil, fmt.Errorf("lexer errors: %v", lexerErrors.errors)
	}
	if len(parserErrors.errors) > 0 {
		return nil, fmt.Errorf("parser errors: %v", parserErrors.errors)
	}

	return &Expression{
		source: expr,
		tree:   tree.(*grammar.EntireExpressionContext),
	}, nil
}
