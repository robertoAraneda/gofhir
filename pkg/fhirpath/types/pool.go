package types

import (
	"sync"
)

// collectionPool is a pool of reusable Collection slices.
var collectionPool = sync.Pool{
	New: func() interface{} {
		// Start with capacity 4, which covers most common cases
		c := make(Collection, 0, 4)
		return &c
	},
}

// GetCollection returns a Collection from the pool.
// The returned collection has length 0 but may have capacity > 0.
func GetCollection() *Collection {
	c := collectionPool.Get().(*Collection)
	return c
}

// PutCollection returns a Collection to the pool for reuse.
// The collection is reset to length 0.
func PutCollection(c *Collection) {
	if c == nil {
		return
	}
	// Clear the slice but keep capacity
	*c = (*c)[:0]
	collectionPool.Put(c)
}

// NewCollectionWithCap creates a new Collection with the specified capacity.
// Use this when you know the expected size to avoid reallocations.
func NewCollectionWithCap(capacity int) Collection {
	return make(Collection, 0, capacity)
}

// SingletonCollection creates a collection with a single value.
// This is a common operation that benefits from optimization.
func SingletonCollection(v Value) Collection {
	return Collection{v}
}

// EmptyCollection is a shared empty collection to avoid allocations.
var EmptyCollection = Collection{}

// booleanPool caches common boolean values to avoid allocations.
var (
	trueBoolean  = Boolean{value: true}
	falseBoolean = Boolean{value: false}
)

// GetBoolean returns a cached Boolean value.
func GetBoolean(b bool) Boolean {
	if b {
		return trueBoolean
	}
	return falseBoolean
}

// TrueCollection is a cached collection containing true.
var TrueCollection = Collection{trueBoolean}

// FalseCollection is a cached collection containing false.
var FalseCollection = Collection{falseBoolean}

// integerCache caches small integers to avoid allocations.
var integerCache [256]Integer

func init() {
	for i := 0; i < 256; i++ {
		integerCache[i] = Integer{value: int64(i - 128)}
	}
}

// GetInteger returns a cached Integer for values in range [-128, 127].
// For other values, creates a new Integer.
func GetInteger(n int64) Integer {
	if n >= -128 && n <= 127 {
		return integerCache[n+128]
	}
	return Integer{value: n}
}
