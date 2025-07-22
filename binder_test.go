package modz

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func newBinderTestFixture(mod Module) (*binder, *assembly) {
	asm, _ := NewAssembly(mod)
	// TODO: Expand Assembly's public API with introspection capabilities so
	// these tests won't need to peek inside the internal implementation.
	internal := asm.(*assembly)
	return newBinder(internal, mod, nil, newModuleSignature(mod)), internal
}

func TestBinder_Install(t *testing.T) {
	mod := &MockModule{
		NameValue: "mod",
		ConfigureFunc: func(binder Binder) error {
			return binder.Install(&MockModule{
				NameValue: "install",
			})
		},
	}
	b, asm := newBinderTestFixture(mod)

	// this will call Install() via mod.ConfigureFunc above.
	err := b.configureModule()
	require.NoError(t, err)

	// check that the second module was added to the assembly.
	require.Len(t, asm.bindings, 2, "should have two module bindings")
}

func TestBinder_getData(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mod",
		ConsumesValue: Keys(ConsumedKey),
		ConfigureFunc: func(binder Binder) error {
			val, err := ConsumedKey.Get(binder)
			require.NoError(t, err)
			require.Equal(t, 42, val)
			return nil
		},
	}
	b, asm := newBinderTestFixture(mod)

	// populate the assembly with a value for consumedKey.
	err := asm.putData(ConsumedKey, 42)
	require.NoError(t, err)

	// getData needs discovery to have run.
	err = b.discoverModule()
	require.NoError(t, err)

	// this will call getData() via mod.ConfigureFunc above.
	err = b.configureModule()
	require.NoError(t, err)
}

func TestBinder_getData_undeclaredKey(t *testing.T) {
	mod := &MockModule{
		NameValue: "mod",
		ConfigureFunc: func(binder Binder) error {
			_, err := ConsumedKey.Get(binder)
			require.Error(t, err)
			return nil
		},
	}
	b, asm := newBinderTestFixture(mod)

	// populate the assembly with a value for consumedKey.
	err := asm.putData(ConsumedKey, 42)
	require.NoError(t, err)

	// getData needs discovery to have run.
	err = b.discoverModule()
	require.NoError(t, err)

	// this will call getData() via mod.ConfigureFunc above.
	err = b.configureModule()
	require.NoError(t, err)
}

func TestBinder_putData(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mod",
		ProducesValue: Keys(ProducedKey),
		ConfigureFunc: func(binder Binder) error {
			err := ProducedKey.Put(binder, "produced")
			require.NoError(t, err)
			return nil
		},
	}
	b, asm := newBinderTestFixture(mod)

	// putData needs discovery to have run.
	err := b.discoverModule()
	require.NoError(t, err)

	// this will call putData() via mod.ConfigureFunc above.
	err = b.configureModule()
	require.NoError(t, err)

	// check that the value was added to the assembly.
	val, err := asm.getData(ProducedKey)
	require.NoError(t, err)
	require.Equal(t, "produced", val)
}

func TestBinder_putData_undeclaredKey(t *testing.T) {
	mod := &MockModule{
		NameValue: "mod",
		ConfigureFunc: func(binder Binder) error {
			err := ProducedKey.Put(binder, "error")
			require.Error(t, err)
			return nil
		},
	}
	b, asm := newBinderTestFixture(mod)

	// putData needs discovery to have run.
	err := b.discoverModule()
	require.NoError(t, err)

	// this will call putData() via mod.ConfigureFunc above.
	err = b.configureModule()
	require.NoError(t, err)

	// check that the value was not added to the assembly.
	_, found := asm.data[ProducedKey]
	require.False(t, found)
}

func TestBinder_discoverModule(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mock",
		ProducesValue: Keys(ProducedKey),
		ConsumesValue: Keys(ConsumedKey),
	}
	b := newBinder(nil, mod, nil, newModuleSignature(mod))

	err := b.discoverModule()
	require.NoError(t, err)
	require.Contains(t, b.produces, ProducedKey)
	require.Contains(t, b.consumes, ConsumedKey)

	require.False(t, b.isReady())
}

func TestBinder_discoverModuleWithNilKeys(t *testing.T) {
	mod := &MockModule{
		NameValue: "mock",
	}
	b := newBinder(nil, mod, nil, newModuleSignature(mod))

	err := b.discoverModule()
	require.NoError(t, err)
	require.Empty(t, b.produces)
	require.Empty(t, b.consumes)

	require.True(t, b.isReady())
}

func TestBinder_discoverModuleWithDuplicateProduces(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mock",
		ProducesValue: Keys(ProducedKey, ProducedKey),
	}
	b := newBinder(nil, mod, nil, newModuleSignature(mod))

	err := b.discoverModule()
	require.Error(t, err)
}

func TestBinder_discoverModuleWithDuplicateConsumes(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mock",
		ConsumesValue: Keys(ConsumedKey, ConsumedKey),
	}
	b := newBinder(nil, mod, nil, newModuleSignature(mod))

	err := b.discoverModule()
	require.Error(t, err)
}

func TestBinder_resolveDependency(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mock",
		ConsumesValue: Keys(ConsumedKey),
	}
	b := newBinder(nil, mod, nil, newModuleSignature(mod))
	err := b.discoverModule()
	require.NoError(t, err)

	require.False(t, b.isReady())
	require.True(t, b.resolveDependency(ConsumedKey))
	require.True(t, b.isReady())
}

func TestBinder_configureModule(t *testing.T) {
	called := false
	mod := &MockModule{
		NameValue: "mod",
		ConfigureFunc: func(binder Binder) error {
			called = true
			return nil
		},
	}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	// this will call mod.ConfigureFunc above.
	err = b.configureModule()
	require.NoError(t, err)
	require.True(t, called)
}

func TestBinder_configureModule_error(t *testing.T) {
	mod := &MockModule{
		NameValue: "error",
		ConfigureFunc: func(binder Binder) error {
			return errors.New("error")
		},
	}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	// this will call mod.ConfigureFunc above.
	err = b.configureModule()
	require.Error(t, err)
}

func TestBinder_configureModule_declaredButNotProduced(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mod",
		ProducesValue: Keys(ProducedKey),
		ConfigureFunc: func(binder Binder) error {
			return nil
		},
	}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	// this will call mod.ConfigureFunc above.
	err = b.configureModule()
	require.Error(t, err)
}
