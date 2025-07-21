package modz

import "fmt"

// Assembly represents a specific composition of [Module]s, defining how they are
// combined to form a complete application or subsystem.
//
// An Assembly is responsible for orchestrating the construction of the dependency graph
// based on the [Data] contracts declared by its modules. It determines the correct
// initialization order for modules and wires up all declared [Data] relationships.
//
// The Assembly manages each [Module]'s lifecycle to build the complete application.
// The Build() method initiates this process by orchestrating each module's
// lifecycle phases to prepare the application for runtime use.
//
// This design keeps the framework focused on construction and wiring, leaving the
// application's runtime behavior and lifecycle management in the hands of the user.
type Assembly interface {
	// Build orchestrates the complete module assembly process.
	//
	// This method orchestrates each [Module]'s lifecycle phases to construct
	// the dependency graph and wire up all module dependencies.
	//
	// Returns an error if any phase fails, such as circular dependencies,
	// missing required [Data], or module configuration errors. On success,
	// the Assembly has completed the construction and wiring phases and
	// is ready for runtime use.
	Build() error

	// sealAssembly is an unexported marker method used to seal the interface.
	sealAssembly()
}

type assembly struct {
	bindings map[moduleSignature]*binder
	data     map[DataKey]any
}

// Ensure that *assembly implements Assembly.
var _ Assembly = (*assembly)(nil)

func (a *assembly) Build() error {
	// TODO: Implement the full build process for the assembly.
	return fmt.Errorf("assembly.Build is not implemented yet")
}

func (*assembly) sealAssembly() {}

// install adds a module into the assembly. Returns an error if the module cannot be installed.
func (a *assembly) install(m Module, parent *binder) error {
	if m == nil {
		return fmt.Errorf("cannot add nil module")
	}
	sig := newModuleSignature(m)
	if _, exists := a.bindings[sig]; exists {
		return fmt.Errorf("module %q already added", sig)
	}
	b := newBinder(a, m, parent, sig)
	a.bindings[sig] = b
	return nil
}

func (a *assembly) getData(key DataKey) (any, error) {
	if key == nil {
		return nil, fmt.Errorf("data key is nil")
	}
	val, ok := a.data[key]
	if !ok {
		return nil, fmt.Errorf("no value for data key")
	}
	return val, nil
}

func (a *assembly) putData(key DataKey, value any) error {
	if key == nil {
		return fmt.Errorf("data key is nil")
	}
	if _, exists := a.data[key]; exists {
		return fmt.Errorf("data key already set")
	}
	a.data[key] = value
	return nil
}

// NewAssembly creates a new Assembly instance with the specified modules.
//
// The provided modules will be included in the Assembly's dependency graph and
// will participate in their lifecycle phases when Build() is called. Modules are
// processed in the order they are provided, though their actual configuration order
// during Build() is determined by their dependency relationships.
//
// Returns an error if the modules cannot be added to the assembly. On success, returns
// an [Assembly] ready for the Build() process.
func NewAssembly(modules ...Module) (Assembly, error) {
	asm := &assembly{
		bindings: make(map[moduleSignature]*binder),
		data:     make(map[DataKey]any),
	}
	for _, m := range modules {
		if err := asm.install(m, nil); err != nil {
			return nil, err
		}
	}
	return asm, nil
}
