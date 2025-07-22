package modz

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAssembly(t *testing.T) {
	m1 := &MockModule{NameValue: "m1"}
	m2 := &MockModule{NameValue: "m2"}
	asm, err := NewAssembly(m1, m2)
	require.NoError(t, err)
	require.NotNil(t, asm)

	internal := asm.(*assembly)
	require.Len(t, internal.bindings, 2, "should have two module bindings")
}

func TestNewAssembly_Duplicate(t *testing.T) {
	m1 := &MockModule{NameValue: "m1"}
	asm, err := NewAssembly(m1, m1)
	require.Error(t, err)
	require.Nil(t, asm)
}

func TestAssembly_Build(t *testing.T) {
	// m1 produces FooKey, m2 consumes FooKey
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			return b.putData(FooKey, 42)
		},
	}
	m2 := &MockModule{
		NameValue:     "m2",
		ConsumesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			v, err := b.getData(FooKey)
			require.NoError(t, err)
			require.Equal(t, 42, v)
			return nil
		},
	}
	asm, err := NewAssembly(m1, m2)
	require.NoError(t, err)
	require.NotNil(t, asm)
	err = asm.Build()
	require.NoError(t, err)
}

func TestAssembly_Build_MissingDependency(t *testing.T) {
	// m1 consumes FooKey, but no module produces it
	m1 := &MockModule{
		NameValue:     "m1",
		ConsumesValue: Keys(FooKey),
	}
	asm, err := NewAssembly(m1)
	require.NoError(t, err)
	require.NotNil(t, asm)
	err = asm.Build()
	require.Error(t, err)
}

func TestAssembly_Build_CircularDependency(t *testing.T) {
	// m1 produces FooKey, consumes BarKey; m2 produces BarKey, consumes FooKey
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(FooKey),
		ConsumesValue: Keys(BarKey),
		ConfigureFunc: func(b Binder) error {
			return b.putData(FooKey, 1)
		},
	}
	m2 := &MockModule{
		NameValue:     "m2",
		ProducesValue: Keys(BarKey),
		ConsumesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			return b.putData(BarKey, 2)
		},
	}
	asm, err := NewAssembly(m1, m2)
	require.NoError(t, err)
	require.NotNil(t, asm)
	err = asm.Build()
	require.Error(t, err)
}

func TestAssembly_Build_ConfigureError(t *testing.T) {
	// m1's Configure returns an error
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			return fmt.Errorf("configure failed")
		},
	}
	asm, err := NewAssembly(m1)
	require.NoError(t, err)
	require.NotNil(t, asm)
	err = asm.Build()
	require.Error(t, err)
}

func TestAssembly_install(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	m := &MockModule{
		NameValue:     "m",
		ProducesValue: Keys(ProducedKey),
		ConsumesValue: Keys(ConsumedKey),
	}
	err := internal.install(m, nil)
	require.NoError(t, err)

	require.Contains(t, internal.waiters, ConsumedKey)
	require.Empty(t, internal.ready)
}

func TestAssembly_install_DependenciesAlreadySatisfied(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)

	// Satisfy ConsumedKey in advance
	err := internal.putData(ConsumedKey, 42)
	require.NoError(t, err)

	m := &MockModule{
		NameValue:     "m",
		ProducesValue: Keys(ProducedKey),
		ConsumesValue: Keys(ConsumedKey),
	}
	err = internal.install(m, nil)
	require.NoError(t, err)

	require.Empty(t, internal.waiters)
	require.Len(t, internal.ready, 1)
}

func TestAssembly_install_Duplicate(t *testing.T) {
	m1 := &MockModule{NameValue: "m1"}
	asm, _ := NewAssembly(m1)
	internal := asm.(*assembly)
	err := internal.install(m1, nil)
	require.Error(t, err)
}

func TestAssembly_install_NilModule(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	err := internal.install(nil, nil)
	require.Error(t, err)
}

func TestAssembly_install_DiscoveryError(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	err := internal.install(&MockModule{
		// Duplicate keys will cause error in discovery
		ProducesValue: Keys(ProducedKey, ProducedKey),
	}, nil)
	require.Error(t, err)
}

func TestAssembly_putData_getData(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)

	// putData: store a value
	err := internal.putData(FooKey, 123)
	require.NoError(t, err)

	// getData: retrieve the value
	val, err := internal.getData(FooKey)
	require.NoError(t, err)
	require.Equal(t, 123, val)
}

func TestAssembly_putData_NilKey(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	err := internal.putData(nil, 1)
	require.Error(t, err)
}

func TestAssembly_putData_DuplicateKey(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)

	// First putData: should succeed
	err := internal.putData(FooKey, 1)
	require.NoError(t, err)

	// Second putData: should fail (duplicate key)
	err = internal.putData(FooKey, 2)
	require.Error(t, err)

	// getData: should return the original value (1)
	val, err := internal.getData(FooKey)
	require.NoError(t, err)
	require.Equal(t, 1, val)
}

func TestAssembly_getData_NilKey(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	_, err := internal.getData(nil)
	require.Error(t, err)
}

func TestAssembly_getData_MissingKey(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	_, err := internal.getData(FooKey)
	require.Error(t, err)
}
