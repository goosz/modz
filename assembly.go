package modz

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Assembly represents a specific composition of [Module], defining how each is
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
// Build can only be called once per Assembly instance; subsequent calls will return an error.
//
// Assembly implements [DataReader], allowing access to the data values produced by modules.
// However, the data access methods (getData) can only be called after Build() has completed
// successfully. Attempting to access data before Build() completes or after Build() fails
// will return an error.
//
// This design keeps the framework focused on construction and wiring, leaving the
// application's runtime behavior and lifecycle management in the hands of the user.
type Assembly interface {
	DataReader

	// Build orchestrates the complete module assembly process.
	//
	// This method orchestrates each [Module]'s lifecycle phases to construct
	// the dependency graph and wire up all module dependencies.
	//
	// Returns an error if any phase fails, such as circular dependencies,
	// missing required [Data], or module configuration errors. On success,
	// the Assembly has completed the construction and wiring phases and
	// is ready for runtime use.
	//
	// Build can only be called once per Assembly instance; subsequent calls will return an error.
	//
	// After Build() completes successfully, the Assembly can be used as a [DataReader]
	// to access the data values produced by modules. Before Build() completes or after
	// Build() fails, data access methods will return an error.
	Build() error

	// sealAssembly is an unexported marker method used to seal the interface.
	sealAssembly()
}

// assembly is the internal implementation of Assembly.
//
// The built field tracks whether Build has already been called, enforcing once-only semantics.
// The buildCompleted field tracks whether Build has completed successfully.
type assembly struct {
	mu             sync.RWMutex // protects all fields below except built and buildCompleted
	bindings       map[moduleSignature]*binder
	registry       *dataRegistry
	data           map[DataKey]any
	waiters        map[DataKey][]*binder
	producers      map[DataKey]*binder // tracks which module produces each data key
	ready          binderQueue
	built          atomic.Bool // true after Build has been called
	buildCompleted atomic.Bool // true after Build has completed successfully
}

// Ensure that *assembly implements Assembly.
var _ Assembly = (*assembly)(nil)

func (a *assembly) Build() error {
	if !a.built.CompareAndSwap(false, true) {
		return fmt.Errorf("Build: can only be called once")
	}
	for {
		a.mu.Lock()
		b := a.ready.Pop()
		a.mu.Unlock()
		if b == nil {
			break
		}
		if err := b.configureModule(); err != nil {
			return err
		}
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.waiters) > 0 {
		// Collect missing keys for error message
		var missingKeys []string
		for k := range a.waiters {
			missingKeys = append(missingKeys, fmt.Sprintf("%v", k))
		}
		return fmt.Errorf("build incomplete: some modules are still waiting for data keys: %v", missingKeys)
	}
	a.buildCompleted.Store(true)
	return nil
}

func (*assembly) sealAssembly() {}

// getData retrieves a value stored under the specified DataKey.
//
// This method can only be called after Build() has completed successfully.
// Returns an error if called before Build() completes or if the DataKey is not found.
func (a *assembly) getData(key DataKey) (any, error) {
	if !a.buildCompleted.Load() {
		return nil, fmt.Errorf("getData: can only be called after Build has completed successfully")
	}
	return a.getDataValue(key)
}

// install adds a module into the assembly. Returns an error if the module cannot be installed.
func (a *assembly) install(m Module, parent *binder) error {
	if m == nil {
		return newInstallError("unknown", "cannot add nil module")
	}
	sig := newModuleSignature(m)
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.bindings[sig]; exists {
		return newInstallError(sig.String(), "already added")
	}
	b := newBinder(a, m, parent, sig)
	if err := b.discoverModule(); err != nil {
		return err
	}

	for k := range b.produces {
		if err := a.registry.Validate(k); err != nil {
			return err
		}
		if existingProducer, exists := a.producers[k]; exists {
			return fmt.Errorf("duplicate producer for data key '%s': modules '%s' and '%s' both declare they produce it", k, existingProducer.moduleSignature, sig)
		}
		a.producers[k] = b
	}

	for k := range b.consumes {
		if err := a.registry.Validate(k); err != nil {
			return err
		}
		if _, present := a.data[k]; !present {
			a.waiters[k] = append(a.waiters[k], b)
		} else {
			b.resolveDependency(k)
		}
	}
	if b.isReady() {
		a.ready.Push(b)
	}
	a.bindings[sig] = b
	return nil
}

// getDataValue retrieves a value from the assembly's data map.
// This is used internally by the binder.
func (a *assembly) getDataValue(key DataKey) (any, error) {
	if key == nil {
		return nil, newDataOperationError(nil, "cannot get data with nil key")
	}
	a.mu.RLock()
	val, ok := a.data[key]
	a.mu.RUnlock()
	if !ok {
		return nil, newDataOperationError(key, "no value found")
	}
	return val, nil
}

// putDataValue stores a value in the assembly's data map and notifies waiters.
// This is used internally by the binder.
func (a *assembly) putDataValue(key DataKey, value any) error {
	if key == nil {
		return newDataOperationError(nil, "cannot put data with nil key")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.data[key]; exists {
		return newDataOperationError(key, "already set")
	}
	a.data[key] = value

	waiters := a.waiters[key]
	if len(waiters) > 0 {
		for _, b := range waiters {
			if b.resolveDependency(key) {
				a.ready.Push(b)
			}
		}
		delete(a.waiters, key)
	}
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
		mu:        sync.RWMutex{},
		bindings:  make(map[moduleSignature]*binder),
		registry:  newDataRegistry(),
		data:      make(map[DataKey]any),
		waiters:   make(map[DataKey][]*binder),
		producers: make(map[DataKey]*binder),
		ready:     make(binderQueue, 0),
	}
	for _, m := range modules {
		if err := asm.install(m, nil); err != nil {
			return nil, err
		}
	}
	return asm, nil
}

// binderQueue is a FIFO queue of *binder.
type binderQueue []*binder

// Push appends a binder to the end of the queue.
func (q *binderQueue) Push(b *binder) {
	*q = append(*q, b)
}

// Pop removes and returns the first binder from the queue, or nil if empty.
func (q *binderQueue) Pop() *binder {
	if len(*q) == 0 {
		return nil
	}
	b := (*q)[0]
	*q = (*q)[1:]
	return b
}
