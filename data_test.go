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
	mock := modz.NewMockDataReadWriter()

	err := fooKey.Put(mock, 42)
	require.NoError(t, err)

	val, err := fooKey.Get(mock)
	require.NoError(t, err)
	require.Equal(t, 42, val)
}

func TestData_Get_TypeMismatch(t *testing.T) {
	mock := modz.NewMockDataReadWriter()
	mock.Store[fooKey] = "not an int"
	_, err := fooKey.Get(mock)
	require.Error(t, err)
}

func TestData_Get_PropagatesReaderError(t *testing.T) {
	mock := modz.NewMockDataReadWriter()
	_, err := fooKey.Get(mock)
	require.Error(t, err, "Get should return any error from the DataReader")
}

func TestData_GetPut_NilInterfaces(t *testing.T) {
	_, err := fooKey.Get(nil)
	require.Error(t, err)
	err = fooKey.Put(nil, 1)
	require.Error(t, err)
}

func TestData_Put_Duplicate(t *testing.T) {
	mock := modz.NewMockDataReadWriter()
	err := fooKey.Put(mock, 1)
	require.NoError(t, err)
	err = fooKey.Put(mock, 2)
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
