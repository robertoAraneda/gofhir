package fhirpath

import (
	"container/list"
	"sync"
	"time"
)

// ExpressionCache provides thread-safe caching of compiled FHIRPath expressions
// with LRU eviction. Use this in production to avoid recompiling the same expressions.
type ExpressionCache struct {
	mu      sync.RWMutex
	cache   map[string]*cacheEntry
	lruList *list.List // Front = most recently used
	limit   int
	hits    int64
	misses  int64
}

type cacheEntry struct {
	expr     *Expression
	key      string
	element  *list.Element
	lastUsed time.Time
}

// CacheStats holds cache performance statistics.
type CacheStats struct {
	Size   int
	Limit  int
	Hits   int64
	Misses int64
}

// NewExpressionCache creates a new cache with the given size limit.
// If limit <= 0, the cache is unbounded.
func NewExpressionCache(limit int) *ExpressionCache {
	return &ExpressionCache{
		cache:   make(map[string]*cacheEntry),
		lruList: list.New(),
		limit:   limit,
	}
}

// Get retrieves a compiled expression from the cache, compiling it if necessary.
func (c *ExpressionCache) Get(expr string) (*Expression, error) {
	// Try read lock first
	c.mu.RLock()
	if entry, ok := c.cache[expr]; ok {
		c.mu.RUnlock()
		// Promote to front (most recently used) - needs write lock
		c.mu.Lock()
		c.lruList.MoveToFront(entry.element)
		entry.lastUsed = time.Now()
		c.hits++
		c.mu.Unlock()
		return entry.expr, nil
	}
	c.mu.RUnlock()

	// Compile the expression
	compiled, err := Compile(expr)
	if err != nil {
		return nil, err
	}

	// Store in cache with write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if entry, ok := c.cache[expr]; ok {
		c.lruList.MoveToFront(entry.element)
		entry.lastUsed = time.Now()
		return entry.expr, nil
	}

	c.misses++

	// LRU eviction if limit reached
	if c.limit > 0 && len(c.cache) >= c.limit {
		c.evictLRU()
	}

	// Create new entry
	entry := &cacheEntry{
		expr:     compiled,
		key:      expr,
		lastUsed: time.Now(),
	}
	entry.element = c.lruList.PushFront(entry)
	c.cache[expr] = entry

	return compiled, nil
}

// evictLRU removes the least recently used entry.
// Must be called with write lock held.
func (c *ExpressionCache) evictLRU() {
	if c.lruList.Len() == 0 {
		return
	}
	// Remove from back (least recently used)
	oldest := c.lruList.Back()
	if oldest != nil {
		entry := oldest.Value.(*cacheEntry)
		c.lruList.Remove(oldest)
		delete(c.cache, entry.key)
	}
}

// MustGet is like Get but panics on error.
func (c *ExpressionCache) MustGet(expr string) *Expression {
	compiled, err := c.Get(expr)
	if err != nil {
		panic(err)
	}
	return compiled
}

// Clear removes all cached expressions.
func (c *ExpressionCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*cacheEntry)
	c.lruList = list.New()
	c.hits = 0
	c.misses = 0
}

// Size returns the number of cached expressions.
func (c *ExpressionCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Stats returns cache performance statistics.
func (c *ExpressionCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStats{
		Size:   len(c.cache),
		Limit:  c.limit,
		Hits:   c.hits,
		Misses: c.misses,
	}
}

// HitRate returns the cache hit rate as a percentage (0-100).
func (c *ExpressionCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total) * 100
}

// DefaultCache is a global expression cache for convenience.
// Use NewExpressionCache for finer control over cache lifetime.
var DefaultCache = NewExpressionCache(1000)

// GetCached retrieves or compiles an expression using the default cache.
func GetCached(expr string) (*Expression, error) {
	return DefaultCache.Get(expr)
}

// MustGetCached is like GetCached but panics on error.
func MustGetCached(expr string) *Expression {
	return DefaultCache.MustGet(expr)
}

// EvaluateCached compiles (with caching) and evaluates a FHIRPath expression.
// This is the recommended function for production use.
func EvaluateCached(resource []byte, expr string) (Collection, error) {
	compiled, err := DefaultCache.Get(expr)
	if err != nil {
		return nil, err
	}
	return compiled.Evaluate(resource)
}
