package funcs

import (
	"context"
	"regexp"
	"sync"
	"time"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
)

// RegexCache provides thread-safe caching of compiled regular expressions
// with LRU eviction and complexity limits for ReDoS protection.
type RegexCache struct {
	mu      sync.RWMutex
	cache   map[string]*regexEntry
	order   []string // LRU order tracking
	limit   int
	maxLen  int           // Maximum regex pattern length
	timeout time.Duration // Default timeout for regex operations
}

type regexEntry struct {
	re       *regexp.Regexp
	lastUsed time.Time
}

// DefaultRegexCache is a global regex cache for production use.
var DefaultRegexCache = NewRegexCache(500, 1000, 100*time.Millisecond)

// NewRegexCache creates a new regex cache with the given parameters.
// - limit: maximum number of cached patterns
// - maxLen: maximum allowed pattern length (ReDoS protection)
// - timeout: default timeout for regex operations
func NewRegexCache(limit, maxLen int, timeout time.Duration) *RegexCache {
	return &RegexCache{
		cache:   make(map[string]*regexEntry),
		order:   make([]string, 0, limit),
		limit:   limit,
		maxLen:  maxLen,
		timeout: timeout,
	}
}

// Compile compiles a regex pattern with caching and complexity validation.
func (c *RegexCache) Compile(pattern string) (*regexp.Regexp, error) {
	// ReDoS protection: check pattern length
	if len(pattern) > c.maxLen {
		return nil, eval.NewEvalError(eval.ErrInvalidExpression,
			"regex pattern too long (max %d characters)", c.maxLen)
	}

	// Check for dangerous patterns (ReDoS prevention)
	if err := validateRegexComplexity(pattern); err != nil {
		return nil, err
	}

	// Try cache first
	c.mu.RLock()
	if entry, ok := c.cache[pattern]; ok {
		entry.lastUsed = time.Now()
		c.mu.RUnlock()
		return entry.re, nil
	}
	c.mu.RUnlock()

	// Compile the pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, eval.NewEvalError(eval.ErrInvalidExpression, "invalid regex: %s", err.Error())
	}

	// Store in cache
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after lock
	if entry, ok := c.cache[pattern]; ok {
		return entry.re, nil
	}

	// LRU eviction if needed
	if len(c.cache) >= c.limit {
		c.evictLRU()
	}

	c.cache[pattern] = &regexEntry{
		re:       re,
		lastUsed: time.Now(),
	}
	c.order = append(c.order, pattern)

	return re, nil
}

// evictLRU removes the least recently used entry.
// Must be called with write lock held.
func (c *RegexCache) evictLRU() {
	if len(c.order) == 0 {
		return
	}

	// Find oldest entry
	oldest := c.order[0]
	oldestIdx := 0
	oldestTime := c.cache[oldest].lastUsed

	for i, pattern := range c.order {
		if entry, ok := c.cache[pattern]; ok {
			if entry.lastUsed.Before(oldestTime) {
				oldest = pattern
				oldestIdx = i
				oldestTime = entry.lastUsed
			}
		}
	}

	// Remove from cache and order
	delete(c.cache, oldest)
	c.order = append(c.order[:oldestIdx], c.order[oldestIdx+1:]...)
}

// MatchWithTimeout performs a regex match with timeout protection.
func (c *RegexCache) MatchWithTimeout(ctx context.Context, pattern, s string) (bool, error) {
	re, err := c.Compile(pattern)
	if err != nil {
		return false, err
	}

	return c.matchWithContext(ctx, re, s)
}

// ReplaceWithTimeout performs a regex replace with timeout protection.
func (c *RegexCache) ReplaceWithTimeout(ctx context.Context, pattern, s, replacement string) (string, error) {
	re, err := c.Compile(pattern)
	if err != nil {
		return "", err
	}

	return c.replaceWithContext(ctx, re, s, replacement)
}

// matchWithContext performs a match with context cancellation checking.
func (c *RegexCache) matchWithContext(ctx context.Context, re *regexp.Regexp, s string) (bool, error) {
	// For short strings, just do the match directly
	if len(s) < 1000 {
		return re.MatchString(s), nil
	}

	// For longer strings, check context periodically
	done := make(chan bool, 1)
	go func() {
		done <- re.MatchString(s)
	}()

	timeout := c.timeout
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < timeout {
			timeout = remaining
		}
	}

	select {
	case result := <-done:
		return result, nil
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(timeout):
		return false, eval.NewEvalError(eval.ErrTimeout, "regex match timeout exceeded")
	}
}

// replaceWithContext performs a replace with context cancellation checking.
func (c *RegexCache) replaceWithContext(ctx context.Context, re *regexp.Regexp, s, replacement string) (string, error) {
	// For short strings, just do the replace directly
	if len(s) < 1000 {
		return re.ReplaceAllString(s, replacement), nil
	}

	// For longer strings, check context periodically
	done := make(chan string, 1)
	go func() {
		done <- re.ReplaceAllString(s, replacement)
	}()

	timeout := c.timeout
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < timeout {
			timeout = remaining
		}
	}

	select {
	case result := <-done:
		return result, nil
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(timeout):
		return "", eval.NewEvalError(eval.ErrTimeout, "regex replace timeout exceeded")
	}
}

// Clear removes all cached patterns.
func (c *RegexCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*regexEntry)
	c.order = make([]string, 0, c.limit)
}

// Size returns the number of cached patterns.
func (c *RegexCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// validateRegexComplexity checks for potentially dangerous regex patterns.
// This helps prevent ReDoS attacks.
func validateRegexComplexity(pattern string) error {
	// Count nested quantifiers and group depth
	var (
		groupDepth     int
		maxGroupDepth  int
		quantifierRun  int
		maxQuantifiers int
		prevWasQuant   bool
	)

	for _, ch := range pattern {
		switch ch {
		case '(':
			groupDepth++
			if groupDepth > maxGroupDepth {
				maxGroupDepth = groupDepth
			}
		case ')':
			if groupDepth > 0 {
				groupDepth--
			}
		case '*', '+', '?':
			quantifierRun++
			if prevWasQuant {
				// Consecutive quantifiers like ** or *+ are dangerous
				return eval.NewEvalError(eval.ErrInvalidExpression,
					"potentially dangerous regex: consecutive quantifiers")
			}
			prevWasQuant = true
		case '{':
			quantifierRun++
			prevWasQuant = true
		default:
			if quantifierRun > maxQuantifiers {
				maxQuantifiers = quantifierRun
			}
			quantifierRun = 0
			prevWasQuant = false
		}
	}

	// Check for excessive nesting (common in ReDoS patterns)
	if maxGroupDepth > 5 {
		return eval.NewEvalError(eval.ErrInvalidExpression,
			"regex has too much nesting (max depth 5)")
	}

	return nil
}
