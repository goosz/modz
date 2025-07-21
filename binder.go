package modz

import (
	"fmt"

	"github.com/goosz/commonz"
)

// Binder represents the controlled interface that modules use to interact with the
// [Assembly] during the module's configuration phase.
//
// The [Assembly] creates a new Binder instance for each module. The Binder acts as a
// facade over the [Assembly], enforcing the framework's access rules while allowing
// modules to wire up their declared dependencies and store their produced values.
//
// The Binder interface is used by modules during their Configure() method calls. Modules
// interact with data through the type-safe [Data][T] interface rather than directly with
// the Binder.
//
// Binder instances should not be retained, shared, or used outside the scope in which
// they are provided to the module. While Binder implementations may be thread-safe,
// they are intended for single-threaded configuration use and should not be used from
// goroutines.
type Binder interface {
	// Install adds a new module to the [Assembly] during the module's configuration phase.
	//
	// This method allows modules to dynamically install additional modules that may be
	// needed based on their configuration or runtime requirements. The newly installed
	// module will go through its own discovery and configuration phases.
	//
	// Returns an error if the module cannot be installed or if called outside of the
	// module's configuration phase.
	Install(Module) error

	// getData retrieves a value stored under the specified DataKey.
	//
	// This method is used internally by the [Data].Get() method to access values from the
	// Assembly's data. The value is returned as [any] and must be type-asserted by the
	// calling code.
	//
	// Returns an error if the [DataKey] is not found, if called outside of the module's
	// configuration phase, or if the calling module did not declare this key in its
	// Consumes() method.
	getData(DataKey) (any, error)

	// putData stores a value under the specified DataKey.
	//
	// This method is used internally by the [Data].Put() method to store values in the
	// Assembly's data. The value is stored as [any].
	//
	// Returns an error if the [DataKey] already has a value stored, if called outside of
	// the module's configuration phase, or if the calling module did not declare this key
	// in its Produces() method.
	putData(DataKey, any) error
}

// TODO: Implement the Binder interface for binder.

// binder is the internal implementation used by [Assembly] to manage a module's lifecycle.
// It encapsulates the state and context for a single module, including its signature, parent binder (if any),
// and a reference to the owning assembly.
//
// Modules interact with the [Binder] interface, not this type directly.
type binder struct {
	moduleSignature moduleSignature
	module          Module
	parent          *binder

	assembly *assembly

	// produces and consumes are populated in the discovery phase.
	produces map[DataKey]struct{}
	consumes map[DataKey]struct{}

	// produced tracks which DataKeys have been produced by this module during configuration.
	produced map[DataKey]struct{}
}

// Ensure that *binder implements Binder.
var _ Binder = (*binder)(nil)

func (b *binder) Install(m Module) error {
	return b.assembly.install(m, b)
}

func (b *binder) getData(key DataKey) (any, error) {
	if _, ok := b.consumes[key]; !ok {
		return nil, fmt.Errorf("module %q did not declare key in Consumes", b.moduleSignature)
	}
	return b.assembly.getData(key)
}

func (b *binder) putData(key DataKey, value any) error {
	if _, ok := b.produces[key]; !ok {
		return fmt.Errorf("module %q did not declare key in Produces", b.moduleSignature)
	}
	err := b.assembly.putData(key, value)
	if err == nil {
		b.produced[key] = struct{}{}
	}
	return err
}

// discoverModule performs the module discovery phase, populating produces and consumes.
func (b *binder) discoverModule() error {
	produces, err := commonz.SliceToSet(b.module.Produces(), true)
	if err != nil {
		return fmt.Errorf("failed to convert produces to set: %w", err)
	}
	consumes, err := commonz.SliceToSet(b.module.Consumes(), true)
	if err != nil {
		return fmt.Errorf("failed to convert consumes to set: %w", err)
	}
	b.produces = produces
	b.consumes = consumes
	return nil
}

// configureModule calls the module's Configure method with this binder and checks all declared produces keys were produced.
func (b *binder) configureModule() error {
	err := b.module.Configure(b)
	if err != nil {
		return err
	}
	// Check that all declared produces keys were actually produced
	var missing []DataKey
	for k := range b.produces {
		if _, ok := b.produced[k]; !ok {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("module %s did not produce all declared keys: %v", b.moduleSignature, missing)
	}
	return nil
}

// newBinder creates a new binder instance for the given module and signature, with an optional parent binder.
// This function is used internally by the [Assembly] to set up the configuration context for a module.
func newBinder(a *assembly, m Module, parent *binder, sig moduleSignature) *binder {
	return &binder{
		moduleSignature: sig,
		module:          m,
		parent:          parent,
		assembly:        a,
		produces:        make(map[DataKey]struct{}),
		consumes:        make(map[DataKey]struct{}),
		produced:        make(map[DataKey]struct{}),
	}
}
