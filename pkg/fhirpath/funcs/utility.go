package funcs

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// TraceLogger defines the interface for structured logging of trace() calls.
type TraceLogger interface {
	Log(entry TraceEntry)
}

// TraceEntry represents a structured trace log entry.
type TraceEntry struct {
	Timestamp  time.Time   `json:"timestamp"`
	Name       string      `json:"name"`
	Input      interface{} `json:"input"`
	Projection interface{} `json:"projection,omitempty"`
	Count      int         `json:"count"`
}

// DefaultTraceLogger logs trace entries to stderr in JSON format.
type DefaultTraceLogger struct {
	mu     sync.Mutex
	writer io.Writer
	json   bool
}

// NewDefaultTraceLogger creates a new default trace logger.
func NewDefaultTraceLogger(writer io.Writer, jsonFormat bool) *DefaultTraceLogger {
	return &DefaultTraceLogger{
		writer: writer,
		json:   jsonFormat,
	}
}

// Log writes a trace entry to the logger's writer.
func (l *DefaultTraceLogger) Log(entry TraceEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.json {
		data, _ := json.Marshal(entry)
		l.writer.Write(data)
		l.writer.Write([]byte("\n"))
	} else {
		if entry.Name != "" {
			io.WriteString(l.writer, "[trace] "+entry.Name+": ")
		} else {
			io.WriteString(l.writer, "[trace] ")
		}
		io.WriteString(l.writer, formatCollection(entry.Input))
		io.WriteString(l.writer, "\n")
		if entry.Projection != nil {
			io.WriteString(l.writer, "[trace] "+entry.Name+" projection: ")
			io.WriteString(l.writer, formatCollection(entry.Projection))
			io.WriteString(l.writer, "\n")
		}
	}
}

// NullTraceLogger discards all trace output (useful for production).
type NullTraceLogger struct{}

// Log does nothing.
func (NullTraceLogger) Log(TraceEntry) {}

// traceLogger is the global trace logger instance.
var (
	traceLogger   TraceLogger = NewDefaultTraceLogger(os.Stderr, false)
	traceLoggerMu sync.RWMutex
)

// SetTraceLogger sets the global trace logger.
// Use NullTraceLogger{} to disable trace output in production.
func SetTraceLogger(logger TraceLogger) {
	traceLoggerMu.Lock()
	defer traceLoggerMu.Unlock()
	traceLogger = logger
}

// GetTraceLogger returns the current trace logger.
func GetTraceLogger() TraceLogger {
	traceLoggerMu.RLock()
	defer traceLoggerMu.RUnlock()
	return traceLogger
}

func formatCollection(input interface{}) string {
	switch v := input.(type) {
	case types.Collection:
		if v.Empty() {
			return "{ }"
		}
		result := "{ "
		for i, item := range v {
			if i > 0 {
				result += ", "
			}
			result += item.String()
		}
		result += " }"
		return result
	default:
		data, _ := json.Marshal(v)
		return string(data)
	}
}

func init() {
	// Register utility functions
	Register(FuncDef{
		Name:    "trace",
		MinArgs: 1,
		MaxArgs: 2,
		Fn:      fnTrace,
	})

	Register(FuncDef{
		Name:    "now",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnNow,
	})

	Register(FuncDef{
		Name:    "today",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToday,
	})

	Register(FuncDef{
		Name:    "timeOfDay",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnTimeOfDay,
	})
}

// fnTrace logs the input collection and returns it unchanged.
// Uses structured logging for production observability.
func fnTrace(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("trace", 1, 0)
	}

	name := ""
	if n, ok := toStringArg(args[0]); ok {
		name = n
	}

	entry := TraceEntry{
		Timestamp: time.Now(),
		Name:      name,
		Input:     collectionToInterface(input),
		Count:     len(input),
	}

	// If a projection is provided, include it
	if len(args) > 1 {
		if result, ok := args[1].(types.Collection); ok {
			entry.Projection = collectionToInterface(result)
		}
	}

	// Log using the configured logger
	GetTraceLogger().Log(entry)

	return input, nil
}

// collectionToInterface converts a Collection to a slice of interface{} for JSON serialization.
func collectionToInterface(col types.Collection) interface{} {
	if col.Empty() {
		return []interface{}{}
	}
	result := make([]interface{}, len(col))
	for i, item := range col {
		result[i] = item.String()
	}
	return result
}

// fnNow returns the current date and time.
func fnNow(_ *eval.Context, _ types.Collection, _ []interface{}) (types.Collection, error) {
	now := time.Now()
	dt, err := types.NewDateTime(now.Format("2006-01-02T15:04:05.000-07:00"))
	if err != nil {
		return types.Collection{}, nil
	}
	return types.Collection{dt}, nil
}

// fnToday returns the current date.
func fnToday(_ *eval.Context, _ types.Collection, _ []interface{}) (types.Collection, error) {
	now := time.Now()
	d, err := types.NewDate(now.Format("2006-01-02"))
	if err != nil {
		return types.Collection{}, nil
	}
	return types.Collection{d}, nil
}

// fnTimeOfDay returns the current time.
func fnTimeOfDay(_ *eval.Context, _ types.Collection, _ []interface{}) (types.Collection, error) {
	now := time.Now()
	t, err := types.NewTime(now.Format("15:04:05.000"))
	if err != nil {
		return types.Collection{}, nil
	}
	return types.Collection{t}, nil
}
