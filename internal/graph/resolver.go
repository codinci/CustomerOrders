package graph

import (
	"errors"
	"sync"
	"time"

	"github.com/codinci/customerorders/internal/graph/scalar"
)

// ErrNotImplemented is returned by resolvers that have not been written yet.
var ErrNotImplemented = errors.New("not implemented")

// Resolver is the root dependency of the GraphQL server.
//
// Phase 0 keeps state in memory. Phase 1 replaces these slices with a
// store.Store interface, so resolvers stop touching storage directly.
//
// Concurrency: gqlgen runs each request on its own goroutine, so every access
// to the fields below must hold mu. Reads take mu.RLock; writes take mu.Lock.
type Resolver struct {
	mu        sync.RWMutex
	customers []customerRecord
	orders    []orderRecord
}

// customerRecord is the stored form of a customer. It holds only IDs, not
// pointers, so records can be copied safely out from under the lock.
type customerRecord struct {
	ID          string
	Name        string
	PhoneNumber string
}

// orderRecord is the stored form of an order. CustomerID links it to its
// customer by ID.
type orderRecord struct {
	ID         string
	CustomerID string
	Product    string
	Quantity   int32
	Price      scalar.Decimal
	CreatedAt  time.Time
}

// NewResolver returns a Resolver with empty in-memory state.
func NewResolver() *Resolver {
	return &Resolver{}
}
