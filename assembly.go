package modz

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
