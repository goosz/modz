package modz

import (
	"fmt"
	"reflect"
	"sync/atomic"

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
	DataKey

	// Get retrieves the value of type T that was stored under this Data key in the provided DataReader.
	// Returns an error if the value is not available or if there is a type mismatch.
	Get(DataReader) (T, error)

	// Put stores a value of type T under this Data key in the provided DataWriter.
	// Returns an error if the DataWriter is nil or if the value cannot be stored.
	Put(DataWriter, T) error
}

// DataKey is a type-erased identifier for a [Data] instance.
type DataKey interface {
	// signature returns the unique identity of the Data key.
	signature() dataKeySignature
}

// DataKeys is a convenience type representing a collection of [DataKey] values.
type DataKeys []DataKey

// Keys is a helper function for constructing a [DataKeys] slice from one or more [DataKey] values.
func Keys(keys ...DataKey) DataKeys {
	return keys
}

// dataKeySignature represents the unique identity of a Data key.
type dataKeySignature struct {
	name string
	pkg  string
}

func (s dataKeySignature) String() string {
	return fmt.Sprintf("%s:%s", s.pkg, s.name)
}

// dataKey is the concrete implementation of the Data interface.
type dataKey[T any] struct {
	dataKeySignature dataKeySignature
	serial           uint64
}

// Ensure that *dataKey[T] implements Data[T].
var _ Data[any] = (*dataKey[any])(nil)

// Global counter for generating unique serial numbers
var dataKeySerialCounter atomic.Uint64

func (d *dataKey[T]) Get(r DataReader) (T, error) {
	if r == nil {
		return commonz.Zero[T](), fmt.Errorf("data reader Get: is nil")
	}
	val, err := r.getData(d)
	if err != nil {
		return commonz.Zero[T](), err
	}
	typedVal, ok := val.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("data key '%v': type assertion failed: expected %T, got %T", d, zero, val)
	}
	return typedVal, nil
}

func (d *dataKey[T]) Put(w DataWriter, t T) error {
	if w == nil {
		return fmt.Errorf("data writer Put: is nil")
	}
	return w.putData(d, t)
}

func (d *dataKey[T]) signature() dataKeySignature {
	return d.dataKeySignature
}

func (d *dataKey[T]) String() string {
	var zero T
	return fmt.Sprintf("Data[%s](%s#%d)", commonz.TypeName(reflect.TypeOf(zero)), d.signature(), d.serial)
}

// NewData creates a new [Data] instance for managing data of type T.
//
// The provided name should be unique within the declaring package and descriptive of the data
// that will be stored under this key. The function automatically captures the package
// information from the calling context to form a unique signature across all packages.
//
// The returned Data key includes package information and a process-unique serial number
// for enhanced identity and debugging.
//
// **Important:** This function must be called from package-level var declarations only.
// It will panic if called from functions, methods, or any other context. This ensures
// proper initialization and prevents runtime conflicts.
func NewData[T any](name string) Data[T] {
	caller := commonz.GetCaller(commonz.ParentCaller)

	if caller.Function != "init" {
		panic(fmt.Sprintf("NewData must be called from package-level var declarations, not from %s.%s", caller.Package, caller.Function))
	}

	serial := dataKeySerialCounter.Add(1)

	return &dataKey[T]{
		dataKeySignature: dataKeySignature{
			name: name,
			pkg:  caller.Package,
		},
		serial: serial,
	}
}
