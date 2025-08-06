package modz

import (
	"fmt"
	"sync"
)

// dataRegistry is a thread-safe registry for validating Data objects.
// It maps dataKeySignature to DataKey and can detect when the same signature
// is presented with different DataKey instances, indicating a clash.
type dataRegistry struct {
	mu    sync.RWMutex
	store map[dataKeySignature]DataKey
}

// newDataRegistry creates a new dataRegistry instance.
func newDataRegistry() *dataRegistry {
	return &dataRegistry{
		store: make(map[dataKeySignature]DataKey),
	}
}

// Validate checks if a DataKey is valid for its signature.
// If the signature is being seen for the first time, it's valid.
// Returns an error if the signature is already registered with a different DataKey.
func (r *dataRegistry) Validate(key DataKey) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sig := key.signature()
	if existing, exists := r.store[sig]; exists {
		if existing != key {
			return fmt.Errorf("data key signature clash: '%s' conflicts with existing key '%s'", sig, existing)
		}
		// Same key, no error
		return nil
	}

	// First time seeing this signature, register it
	r.store[sig] = key
	return nil
}
