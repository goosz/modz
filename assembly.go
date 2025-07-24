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
	//
	// Build can only be called once per Assembly instance; subsequent calls will return an error.
	Build() error

	// sealAssembly is an unexported marker method used to seal the interface.
	sealAssembly()
}

// assembly is the internal implementation of Assembly.
//
// The built field tracks whether Build has already been called, enforcing once-only semantics.
type assembly struct {
	mu       sync.RWMutex // protects all fields below except built
	bindings map[moduleSignature]*binder
	data     map[DataKey]any
	waiters  map[DataKey][]*binder
	ready    binderQueue
	built    atomic.Bool // true after Build has been called
}

// Ensure that *assembly implements Assembly.
var _ Assembly = (*assembly)(nil)

func (a *assembly) Build() error {
	if !a.built.CompareAndSwap(false, true) {
		return fmt.Errorf("Build can only be called once on this Assembly")
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
	return nil
}

func (*assembly) sealAssembly() {}

// install adds a module into the assembly. Returns an error if the module cannot be installed.
func (a *assembly) install(m Module, parent *binder) error {
	if m == nil {
		return fmt.Errorf("cannot add nil module")
	}
	sig := newModuleSignature(m)
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.bindings[sig]; exists {
		return fmt.Errorf("module %q already added", sig)
	}
	b := newBinder(a, m, parent, sig)
	if err := b.discoverModule(); err != nil {
		return err
	}
	for k := range b.consumes {
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

func (a *assembly) getData(key DataKey) (any, error) {
	if key == nil {
		return nil, fmt.Errorf("data key is nil")
	}
	a.mu.RLock()
	val, ok := a.data[key]
	a.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no value for data key")
	}
	return val, nil
}

func (a *assembly) putData(key DataKey, value any) error {
	if key == nil {
		return fmt.Errorf("data key is nil")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.data[key]; exists {
		return fmt.Errorf("data key already set")
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
		mu:       sync.RWMutex{},
		bindings: make(map[moduleSignature]*binder),
		data:     make(map[DataKey]any),
		waiters:  make(map[DataKey][]*binder),
		ready:    make(binderQueue, 0),
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
