package modz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test keys for registry testing
var (
	registryTestKey1 = NewData[int]("registry-test-1")
	registryTestKey2 = NewData[int]("registry-test-2")
	registryTestKey3 = NewData[string]("registry-test-3") // Different name, different type
)

func TestDataRegistry_Validate(t *testing.T) {
	registry := newDataRegistry()

	// First time seeing this signature - should register it
	err := registry.Validate(registryTestKey1)
	require.NoError(t, err)

	// Same key again - should not error
	err = registry.Validate(registryTestKey1)
	require.NoError(t, err)

	// Different key - should not error
	err = registry.Validate(registryTestKey2)
	require.NoError(t, err)
}

func TestDataRegistry_ValidateSignatureClash(t *testing.T) {
	registry := newDataRegistry()

	// Register the first key
	err := registry.Validate(ClashTestKey1)
	require.NoError(t, err)

	// Try to register the second key with the same signature
	err = registry.Validate(ClashTestKey2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "data key signature clash")
}

func TestDataRegistry_ValidateDifferentTypes(t *testing.T) {
	registry := newDataRegistry()

	// Both should be valid since they have different signatures (different names)
	err := registry.Validate(registryTestKey1)
	require.NoError(t, err)

	err = registry.Validate(registryTestKey3)
	require.NoError(t, err)
}
