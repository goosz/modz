package modz

import (
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

func TestAssembly_install(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	m := &MockModule{NameValue: "m"}
	err := internal.install(m, nil)
	require.NoError(t, err)
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
