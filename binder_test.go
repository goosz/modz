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

func TestBinder_Install_outsideConfigPhase(t *testing.T) {
	mod := &MockModule{NameValue: "mod"}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	// Install should fail outside of configuration phase
	err = b.Install(&MockModule{NameValue: "other"})
	require.Error(t, err)
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
	err := asm.putDataValue(ConsumedKey, 42)
	require.NoError(t, err)

	// getData needs discovery to have run.
	err = b.discoverModule()
	require.NoError(t, err)

	// this will call getData() via mod.ConfigureFunc above.
	err = b.configureModule()
	require.NoError(t, err)
}

func TestBinder_getData_outsideConfigPhase(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mod",
		ConsumesValue: Keys(ConsumedKey),
	}
	b, asm := newBinderTestFixture(mod)
	err := asm.putDataValue(ConsumedKey, 42)
	require.NoError(t, err)
	err = b.discoverModule()
	require.NoError(t, err)

	// getData should fail outside of configuration phase
	_, err = b.getData(ConsumedKey)
	require.Error(t, err)
}

func TestBinder_getData_undeclaredKey(t *testing.T) {
	mod := &MockModule{
		NameValue: "mod",
		ConfigureFunc: func(binder Binder) error {
			_, err := ConsumedKey.Get(binder)
			require.Error(t, err)
			return err
		},
	}
	b, asm := newBinderTestFixture(mod)

	// populate the assembly with a value for consumedKey.
	err := asm.putDataValue(ConsumedKey, 42)
	require.NoError(t, err)

	// getData needs discovery to have run.
	err = b.discoverModule()
	require.NoError(t, err)

	// this will call getData() via mod.ConfigureFunc above.
	err = b.configureModule()
	require.Error(t, err)
}

func TestBinder_getData_assemblyError(t *testing.T) {
	// Test getData when assembly.getDataValue returns an error
	mod := &MockModule{
		NameValue:     "test",
		ConsumesValue: Keys(ConsumedKey),
		ConfigureFunc: func(b Binder) error {
			// This should fail because the assembly doesn't have the value
			_, err := b.getData(ConsumedKey)
			return err
		},
	}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	err = b.configureModule()
	require.Error(t, err)

	// Verify it's a ConfigurationError with proper context
	var configErr *ConfigurationError
	require.ErrorAs(t, err, &configErr)
	require.Equal(t, "test", configErr.ModuleName)
	require.Equal(t, "getData", configErr.Operation)
	require.Contains(t, configErr.Error(), "data key 'Data[int](github.com/goosz/modz:consumed#5)': no value found")
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
	val, err := asm.getDataValue(ProducedKey)
	require.NoError(t, err)
	require.Equal(t, "produced", val)
}

func TestBinder_putData_outsideConfigPhase(t *testing.T) {
	mod := &MockModule{
		NameValue:     "mod",
		ProducesValue: Keys(ProducedKey),
	}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	// putData should fail outside of configuration phase
	err = b.putData(ProducedKey, "value")
	require.Error(t, err)
}

func TestBinder_putData_undeclaredKey(t *testing.T) {
	mod := &MockModule{
		NameValue: "mod",
		ConfigureFunc: func(binder Binder) error {
			err := ProducedKey.Put(binder, "error")
			require.Error(t, err)
			return err
		},
	}
	b, asm := newBinderTestFixture(mod)

	// putData needs discovery to have run.
	err := b.discoverModule()
	require.NoError(t, err)

	// this will call putData() via mod.ConfigureFunc above.
	err = b.configureModule()
	require.Error(t, err)

	// check that the value was not added to the assembly.
	_, found := asm.data[ProducedKey]
	require.False(t, found)
}

func TestBinder_putData_assemblyError(t *testing.T) {
	// Test putData when assembly.putDataValue returns an error
	mod := &MockModule{
		NameValue:     "test",
		ProducesValue: Keys(ProducedKey),
		ConfigureFunc: func(b Binder) error {
			// First put should succeed
			err := b.putData(ProducedKey, "first")
			if err != nil {
				return err
			}
			// Second put should fail (duplicate key)
			err = b.putData(ProducedKey, "second")
			return err
		},
	}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	err = b.configureModule()
	require.Error(t, err)

	// Verify it's a ConfigurationError with proper context
	var configErr *ConfigurationError
	require.ErrorAs(t, err, &configErr)
	require.Equal(t, "test", configErr.ModuleName)
	require.Equal(t, "putData", configErr.Operation)
	require.Contains(t, configErr.Error(), "data key 'Data[string](github.com/goosz/modz:produced#4)': already set")
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

	// Test that no configuration errors are tracked when configuration succeeds
	trackedError := b.GetConfigurationError()
	require.Nil(t, trackedError)
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

	// Verify it's a ConfigurationError with proper context
	var configErr *ConfigurationError
	require.ErrorAs(t, err, &configErr)
	require.Equal(t, "error", configErr.ModuleName)
	require.Equal(t, "Configure", configErr.Operation)
	require.Contains(t, configErr.Error(), "error")

	// Test error tracking
	trackedError := b.GetConfigurationError()
	require.NotNil(t, trackedError)
	require.Contains(t, trackedError.Error(), "error")
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

	// Verify it's a ConfigurationError with proper context
	var configErr *ConfigurationError
	require.ErrorAs(t, err, &configErr)
	require.Equal(t, "mod", configErr.ModuleName)
	require.Equal(t, "Configure", configErr.Operation)
	require.Contains(t, configErr.Error(), "module did not produce all declared keys")

	// Test error tracking
	trackedError := b.GetConfigurationError()
	require.NotNil(t, trackedError)
	require.Contains(t, trackedError.Error(), "module did not produce all declared keys")
}

func TestBinder_configureModule_twice(t *testing.T) {
	mod := &MockModule{
		NameValue: "mod",
		ConfigureFunc: func(binder Binder) error {
			return nil
		},
	}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	err = b.configureModule()
	require.NoError(t, err)

	err = b.configureModule()
	require.Error(t, err, "second call to configureModule should return an error")
}

func TestBinder_configureModule_errorSwallowing(t *testing.T) {
	badModule := &MockModule{
		NameValue: "BadModule",
		ConfigureFunc: func(b Binder) error {
			// This is BAD: we're calling Install but ignoring the error
			_ = b.Install(&MockModule{NameValue: "BadModule"}) // This will fail (duplicate module) but we ignore it

			// This is also BAD: we're trying to get an undeclared key but ignoring the error
			_, _ = ConsumedKey.Get(b) // This will fail but we ignore it

			// This is also BAD: we're trying to put an undeclared key but ignoring the error
			_ = ProducedKey.Put(b, "oops") // This will fail but we ignore it

			// Return nil despite encountering errors - this is what we want to detect
			return nil
		},
	}
	assembly, err := NewAssembly(badModule)
	require.NoError(t, err)

	err = assembly.Build()
	require.Error(t, err)

	// Verify we get a ConfigurationError with proper context
	var moduleErr *ConfigurationError
	require.ErrorAs(t, err, &moduleErr)
	require.Equal(t, "BadModule", moduleErr.ModuleName)
	require.Equal(t, "Install", moduleErr.Operation)
	require.Contains(t, moduleErr.Error(), "module 'BadModule': already added")
}

func TestConfigurationError_Error_WithNilErr(t *testing.T) {
	configErr := &ConfigurationError{
		ModuleName: "TestModule",
		Operation:  "TestOperation",
		Err:        nil,
	}

	errorMsg := configErr.Error()
	require.Contains(t, errorMsg, "TestModule")
	require.Contains(t, errorMsg, "TestOperation")
	require.Contains(t, errorMsg, "failed")
}

func TestBinder_failFastBehavior(t *testing.T) {
	// Test that operations fail fast when there's already an error
	mod := &MockModule{
		NameValue:     "test",
		ProducesValue: Keys(ProducedKey),
		ConsumesValue: Keys(ConsumedKey),
		ConfigureFunc: func(b Binder) error {
			// First operation that will fail
			err := b.putData(ProducedKey, "first")
			if err != nil {
				return err
			}

			// Second put should fail (duplicate key) and set the error
			err = b.putData(ProducedKey, "second")
			if err != nil {
				return err
			}

			// These operations should all fail fast with the first error
			err = b.Install(&MockModule{NameValue: "should-fail-fast"})
			if err != nil {
				return err
			}

			_, err = b.getData(ConsumedKey)
			if err != nil {
				return err
			}

			err = b.putData(ProducedKey, "should-also-fail-fast")
			return err
		},
	}
	b, _ := newBinderTestFixture(mod)
	err := b.discoverModule()
	require.NoError(t, err)

	err = b.configureModule()
	require.Error(t, err)

	// Verify it's a ConfigurationError with the first error
	var configErr *ConfigurationError
	require.ErrorAs(t, err, &configErr)
	require.Equal(t, "test", configErr.ModuleName)
	require.Equal(t, "putData", configErr.Operation)
	require.Contains(t, configErr.Error(), "data key 'Data[string](github.com/goosz/modz:produced#4)': already set")

	// Verify the tracked error is the first one
	trackedError := b.GetConfigurationError()
	require.NotNil(t, trackedError)
	require.Contains(t, trackedError.Error(), "data key 'Data[string](github.com/goosz/modz:produced#4)': already set")
}
