package modz

import (
	"errors"
	"fmt"

	"github.com/goosz/commonz"
)

// DataReader defines the interface for reading data values from a storage mechanism.
type DataReader interface {
	// getData retrieves a value stored under the specified DataKey.
	//
	// This method is used internally by the [Data].Get() method to access values.
	// The value is returned as [any] and must be type-asserted by the calling code.
	//
	// Returns an error if the [DataKey] is not found or if the operation is not allowed.
	getData(DataKey) (any, error)
}

// DataWriter defines the interface for writing data values to a storage mechanism.
type DataWriter interface {
	// putData stores a value under the specified DataKey.
	//
	// This method is used internally by the [Data].Put() method to store values.
	// The value is stored as [any].
	//
	// Returns an error if the [DataKey] already has a value stored or if the operation is not allowed.
	putData(DataKey, any) error
}

// Data represents a type-safe key for dependency injection within a Modz application.
//
// Each Data instance acts as both a key (for storing/retrieving values) and a contract
// (defining the type and purpose) for a value of type T that can be produced by one module
// and consumed by others.
//
// During the module's discovery phase, modules declare their dependencies using Data keys via
// Produces() and Consumes() methods.
//
// During the module's configuration phase, the Assembly provides modules with access to dependencies
// through data access interfaces (typically via the Binder), allowing modules to wire up their
// required values using Get() and Put() methods on their declared Data keys.
//
// Data keys are intended to be used by application and module authors, not implemented directly.
// Always use [NewData] to create new Data keys.
type Data[T any] interface {
	// Get retrieves the value of type T that was stored under this Data key in the provided DataReader.
	// Returns an error if the value is not available or if there is a type mismatch.
	Get(DataReader) (T, error)

	// Put stores a value of type T under this Data key in the provided DataWriter.
	// Returns an error if the DataWriter is nil or if the value cannot be stored.
	Put(DataWriter, T) error

	// sealData is an unexported marker method used to seal the interface.
	sealData()
}

// DataKey is a type-erased identifier for a [Data] instance.
type DataKey interface {
	// sealData is an unexported marker method used to seal the interface.
	sealData()
}

// Ensure that Data[T] implements DataKey.
var _ DataKey = (Data[any])(nil)

// DataKeys is a convenience type representing a collection of [DataKey] values.
type DataKeys []DataKey

// Keys is a helper function for constructing a [DataKeys] slice from one or more [DataKey] values.
func Keys(keys ...DataKey) DataKeys {
	return keys
}

// dataKey is the concrete implementation of the Data interface.
type dataKey[T any] struct {
	name string
}

// Ensure that *dataKey[T] implements Data[T].
var _ Data[any] = (*dataKey[any])(nil)

func (d *dataKey[T]) Get(r DataReader) (T, error) {
	if r == nil {
		return commonz.Zero[T](), errors.New("data reader is nil")
	}
	val, err := r.getData(d)
	if err != nil {
		return commonz.Zero[T](), err
	}
	typedVal, ok := val.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("type assertion failed: expected %T, got %T", zero, val)
	}
	return typedVal, nil
}

func (d *dataKey[T]) Put(w DataWriter, t T) error {
	if w == nil {
		return errors.New("data writer is nil")
	}
	return w.putData(d, t)
}

func (d *dataKey[T]) sealData() {}

func (d *dataKey[T]) String() string {
	return fmt.Sprintf("Data[%T](%q)", *new(T), d.name)
}

// NewData creates a new [Data] instance for managing data of type T.
//
// The provided name should be unique within your application and descriptive of the data
// that will be stored under this key. While uniqueness is not currently enforced, it is
// strongly recommended to avoid conflicts and to aid in debugging or logging.
//
// Note: This implementation of NewData is temporary and incomplete. Additional features and
// enforcement may be added as the framework evolves.
func NewData[T any](name string) Data[T] {
	return &dataKey[T]{name: name}
}
