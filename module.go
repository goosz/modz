package modz

import (
	"reflect"
)

// Module represents a modular component within a Modz application.
//
// Each module declares what [Data] it produces and consumes, and provides a configuration
// method to wire up its dependencies. Modules are the building blocks of modular
// applications, allowing for clear separation of concerns and loose coupling.
//
// Each module goes through two phases in its lifecycle:
//   - Discovery Phase: The [Assembly] calls Produces() and Consumes() to understand
//     what [Data] the module provides and requires. This builds the dependency graph.
//   - Configuration Phase: The [Assembly] calls Configure() after all of the
//     module's consumed [Data] has been produced by other modules. During this phase, the
//     module wires up its dependencies and stores its produced values.
//
// The Name(), Produces(), and Consumes() methods must have deterministic behavior:
// they must return the same values when called repeatedly on the same module instance.
// While modules may have dynamic behavior based on their construction parameters,
// these three methods must be consistent across multiple calls.
//
// Additionally, the Configure() method must only interact with [DataKey] values that were
// declared in Produces() and Consumes(). A module cannot Put() to a [DataKey] it
// did not declare in Produces(), nor Get() from a [DataKey] it did not declare in
// Consumes(). This ensures consistency between the module's discovery and configuration phases.
type Module interface {
	// Name returns the identifier for this module.
	//
	// The name should be descriptive and unique within the module's package to aid
	// in debugging, logging, and dependency resolution. Module names are combined
	// with the package path to form a unique signature across all packages.
	// This allows different packages to have modules with the same name without conflicts.
	// Module signatures are used by the [Assembly] for error reporting and dependency graph visualization.
	Name() string

	// Produces returns the set of [DataKey]s that this module provides.
	//
	// During the module's discovery phase, the [Assembly] calls this method to understand
	// what [Data] this module makes available to other modules. The returned
	// [DataKeys] represent the contracts that this module fulfills.
	Produces() DataKeys

	// Consumes returns the [DataKey]s that this module requires.
	//
	// During the module's discovery phase, the [Assembly] calls this method to understand
	// what [Data] this module depends on. The returned [DataKeys] represent the
	// contracts that must be fulfilled by other modules before this module
	// can be configured.
	Consumes() DataKeys

	// Configure wires up the module's dependencies using the provided [Binder].
	//
	// Called during the module's configuration phase. The module should use the [Binder]
	// to retrieve its required dependencies via Get() calls on its consumed [DataKey]s,
	// and store any values it produces via Put() calls on its produced [DataKey]s.
	//
	// The module may also install additional modules via the [Binder] if needed.
	//
	// The provided [Binder] is specific to this module instance and should not
	// be retained, shared, or used outside the scope of this Configure() call.
	//
	// Configure should be fast to execute and should not perform any heavy work
	// such as starting services, opening connections, or loading large amounts
	// of data. Such initialization should be deferred to runtime after the
	// [Assembly] Build() has completed.
	//
	// **Error Handling Requirements:**
	// - Configure MUST return any errors encountered from Binder operations (Install, Get, Put)
	// - Do not swallow or ignore errors from Binder methods
	// - The framework will detect and report modules that return nil despite encountering errors
	// - Return errors immediately when they occur; do not continue with partial configuration
	//
	// Returns an error if configuration fails, which will halt the [Assembly] build
	// process.
	Configure(Binder) error
}

// Singleton is a marker interface that can be embedded in modules to indicate
// they will be silently ignored when installed multiple times. All modules
// can only be installed once per assembly, but singleton modules won't
// return an error on subsequent installation attempts.
type Singleton struct{}

// singleton is an unexported marker method that identifies singleton modules.
func (s *Singleton) singleton() {
	// Marker method - no implementation needed
}

// moduleSignature uniquely identifies a Module within an Assembly.
//
// It is used as a key in internal maps to track module bindings and ensure uniqueness.
// The signature derives from the Module's package and Name() method,
// allowing different packages to have modules with the same name without conflicts.
type moduleSignature struct {
	packageName string
	name        string
}

func (sig moduleSignature) String() string {
	return sig.packageName + ":" + sig.name
}

// newModuleSignature creates a new moduleSignature for the given Module.
func newModuleSignature(m Module) moduleSignature {
	return moduleSignature{
		packageName: reflect.TypeOf(m).Elem().PkgPath(),
		name:        m.Name(),
	}
}
