package modz_test

import (
	"fmt"
	"testing"

	"github.com/goosz/modz"
	"github.com/stretchr/testify/require"
)

// Package-level Data keys for tests
var (
	fooKey = modz.NewData[int]("foo")
	barKey = modz.NewData[int]("bar")
)

func TestData_PutAndGet(t *testing.T) {
	binder := modz.NewMockBinder()
	err := fooKey.Put(binder, 42)
	require.NoError(t, err)
	val, err := fooKey.Get(binder)
	require.NoError(t, err)
	require.Equal(t, 42, val)
}

func TestData_Get_TypeMismatch(t *testing.T) {
	binder := modz.NewMockBinder()
	binder.Store[fooKey] = "not an int"
	_, err := fooKey.Get(binder)
	require.Error(t, err)
}

func TestData_Get_PropagatesBinderError(t *testing.T) {
	binder := modz.NewMockBinder()
	_, err := fooKey.Get(binder)
	require.Error(t, err, "Get should return any error from the Binder")
}

func TestData_GetPut_NilBinder(t *testing.T) {
	_, err := fooKey.Get(nil)
	require.Error(t, err)
	err = fooKey.Put(nil, 1)
	require.Error(t, err)
}

func TestData_Put_Duplicate(t *testing.T) {
	binder := modz.NewMockBinder()
	err := fooKey.Put(binder, 1)
	require.NoError(t, err)
	err = fooKey.Put(binder, 2)
	require.Error(t, err, "Put should return an error if the key is already set")
}

func TestKeysHelper(t *testing.T) {
	keys := modz.Keys(barKey, fooKey)
	require.Equal(t, modz.DataKeys{barKey, fooKey}, keys)
}

func TestDataKey_String(t *testing.T) {
	str := fmt.Sprintf("%v", fooKey)
	require.Contains(t, str, "foo", "String() should include the key name")
	require.Contains(t, str, "int", "String() should include the type name")
}
