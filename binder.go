package modz

import (
	"fmt"

	"sync/atomic"

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
//
// The framework strictly enforces that Install and the data access methods (getData and putData) may only be called during the
// configuration phase (i.e., while the module's Configure method is running). If these methods
// are called outside of this phase, they will return an error.
//
// Additionally, configureModule can only be called once per binder; subsequent calls will return an error.
type Binder interface {
	DataReader
	DataWriter

	// Install adds a new module to the [Assembly] during the module's configuration phase.
	//
	// This method allows modules to dynamically install additional modules that may be
	// needed based on their configuration or runtime requirements. The newly installed
	// module will go through its own discovery and configuration phases.
	//
	// Returns an error if the module cannot be installed or if called outside of the
	// module's configuration phase (strictly enforced).
	Install(Module) error
}

// binder is the internal implementation used by [Assembly] to manage a module's lifecycle.
// It encapsulates the state and context for a single module, including its signature, parent binder (if any),
// and a reference to the owning assembly.
//
// Modules interact with the [Binder] interface, not this type directly.
//
// The binder enforces once-only semantics for configureModule (it can only be called once per binder),
// and strictly enforces that Install and the data access methods (getData and putData) may only be called
// during the configuration phase. Errors are returned if these rules are violated.
type binder struct {
	moduleSignature moduleSignature
	module          Module
	parent          *binder

	assembly *assembly

	// produces and consumes are populated in the discovery phase.
	produces map[DataKey]struct{}
	consumes map[DataKey]struct{}

	// waiting contains DataKeys waiting to be satisfied before this module's configuration can begin.
	waiting map[DataKey]struct{}

	// produced tracks which DataKeys have been produced by this module during configuration.
	produced map[DataKey]struct{}

	// inProgress is true while configureModule is running.
	inProgress atomic.Bool
	// configured is true after configureModule has run once.
	configured atomic.Bool
}

// Ensure that *binder implements Binder and the data interfaces.
var _ Binder = (*binder)(nil)
var _ DataReader = (*binder)(nil)
var _ DataWriter = (*binder)(nil)

func (b *binder) Install(m Module) error {
	if !b.inProgress.Load() {
		return fmt.Errorf("Install can only be called during module configuration phase for module %q", b.moduleSignature)
	}
	return b.assembly.install(m, b)
}

func (b *binder) getData(key DataKey) (any, error) {
	if !b.inProgress.Load() {
		return nil, fmt.Errorf("getData can only be called during module configuration phase for module %q", b.moduleSignature)
	}
	if _, ok := b.consumes[key]; !ok {
		return nil, fmt.Errorf("module %q did not declare key in Consumes", b.moduleSignature)
	}
	return b.assembly.getDataValue(key)
}

func (b *binder) putData(key DataKey, value any) error {
	if !b.inProgress.Load() {
		return fmt.Errorf("putData can only be called during module configuration phase for module %q", b.moduleSignature)
	}
	if _, ok := b.produces[key]; !ok {
		return fmt.Errorf("module %q did not declare key in Produces", b.moduleSignature)
	}
	err := b.assembly.putDataValue(key, value)
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
	// initialize waiting as a copy of consumes
	for k := range consumes {
		b.waiting[k] = struct{}{}
	}
	return nil
}

// isReady reports whether all dependencies for this module have been satisfied and it is ready to be configured.
func (b *binder) isReady() bool {
	return len(b.waiting) == 0
}

// resolveDependency marks the given DataKey as satisfied and returns true if all dependencies are now satisfied.
func (b *binder) resolveDependency(k DataKey) bool {
	delete(b.waiting, k)
	return len(b.waiting) == 0
}

// configureModule calls the module's Configure method with this binder and checks all declared produces keys were produced.
// It can only be called once; subsequent calls return an error.
func (b *binder) configureModule() error {
	if !b.configured.CompareAndSwap(false, true) {
		return fmt.Errorf("configureModule can only be called once for module %q", b.moduleSignature)
	}
	b.inProgress.Store(true)
	err := b.module.Configure(b)
	b.inProgress.Store(false)
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
		waiting:         make(map[DataKey]struct{}),
		produced:        make(map[DataKey]struct{}),
	}
}
