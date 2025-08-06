package modz

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAssembly(t *testing.T) {
	m1 := &MockModule{NameValue: "m1"}
	m2 := &MockModule{NameValue: "m2"}
	asm, err := NewAssembly(m1, m2)
	require.NoError(t, err)
	require.NotNil(t, asm)

	internal := asm.(*assembly)
	require.Len(t, internal.bindings, 2, "should have two module bindings")
}

func TestNewAssembly_Duplicate(t *testing.T) {
	m1 := &MockModule{NameValue: "m1"}
	asm, err := NewAssembly(m1, m1)
	require.Error(t, err)
	require.Nil(t, asm)
}

func TestAssembly_Build(t *testing.T) {
	// m1 produces FooKey, m2 consumes FooKey
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			return b.putData(FooKey, 42)
		},
	}
	m2 := &MockModule{
		NameValue:     "m2",
		ConsumesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			v, err := b.getData(FooKey)
			require.NoError(t, err)
			require.Equal(t, 42, v)
			return nil
		},
	}
	asm, err := NewAssembly(m1, m2)
	require.NoError(t, err)
	require.NotNil(t, asm)
	err = asm.Build()
	require.NoError(t, err)
}

func TestAssembly_Build_MissingDependency(t *testing.T) {
	// m1 consumes FooKey, but no module produces it
	m1 := &MockModule{
		NameValue:     "m1",
		ConsumesValue: Keys(FooKey),
	}
	asm, err := NewAssembly(m1)
	require.NoError(t, err)
	require.NotNil(t, asm)
	err = asm.Build()
	require.Error(t, err)
}

func TestAssembly_Build_CircularDependency(t *testing.T) {
	// m1 produces FooKey, consumes BarKey; m2 produces BarKey, consumes FooKey
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(FooKey),
		ConsumesValue: Keys(BarKey),
		ConfigureFunc: func(b Binder) error {
			return b.putData(FooKey, 1)
		},
	}
	m2 := &MockModule{
		NameValue:     "m2",
		ProducesValue: Keys(BarKey),
		ConsumesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			return b.putData(BarKey, 2)
		},
	}
	asm, err := NewAssembly(m1, m2)
	require.NoError(t, err)
	require.NotNil(t, asm)
	err = asm.Build()
	require.Error(t, err)
}

func TestAssembly_Build_ConfigureError(t *testing.T) {
	// m1's Configure returns an error
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			return fmt.Errorf("configure failed")
		},
	}
	asm, err := NewAssembly(m1)
	require.NoError(t, err)
	require.NotNil(t, asm)
	err = asm.Build()
	require.Error(t, err)

	// Verify it's a ConfigurationError with proper context
	var configErr *ConfigurationError
	require.ErrorAs(t, err, &configErr)
	require.Equal(t, "github.com/goosz/modz:m1", configErr.ModuleID)
	require.Equal(t, "Configure", configErr.Operation)
	require.Contains(t, configErr.Error(), "configure failed")
}

func TestAssembly_Build_Twice(t *testing.T) {
	m := &MockModule{NameValue: "m"}
	asm, err := NewAssembly(m)
	require.NoError(t, err)
	require.NotNil(t, asm)

	err = asm.Build()
	require.NoError(t, err)

	err = asm.Build()
	require.Error(t, err, "second call to Build should return an error")
}

func TestAssembly_install(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	m := &MockModule{
		NameValue:     "m",
		ProducesValue: Keys(ProducedKey),
		ConsumesValue: Keys(ConsumedKey),
	}
	err := internal.install(m, nil)
	require.NoError(t, err)

	require.Contains(t, internal.waiters, ConsumedKey)
	require.Empty(t, internal.ready)
}

func TestAssembly_install_DependenciesAlreadySatisfied(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)

	// Satisfy ConsumedKey in advance
	err := internal.putDataValue(ConsumedKey, 42)
	require.NoError(t, err)

	m := &MockModule{
		NameValue:     "m",
		ProducesValue: Keys(ProducedKey),
		ConsumesValue: Keys(ConsumedKey),
	}
	err = internal.install(m, nil)
	require.NoError(t, err)

	require.Empty(t, internal.waiters)
	require.Len(t, internal.ready, 1)
}

func TestAssembly_install_Duplicate(t *testing.T) {
	m1 := &MockModule{NameValue: "m1"}
	asm, _ := NewAssembly(m1)
	internal := asm.(*assembly)
	err := internal.install(m1, nil)
	require.Error(t, err)
}

func TestAssembly_install_NilModule(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	err := internal.install(nil, nil)
	require.Error(t, err)
}

func TestAssembly_install_DiscoveryError(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	err := internal.install(&MockModule{
		// Duplicate keys will cause error in discovery
		ProducesValue: Keys(ProducedKey, ProducedKey),
	}, nil)
	require.Error(t, err)
}

func TestAssembly_putDataValue_getDataValue(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)

	// putDataValue: store a value
	err := internal.putDataValue(FooKey, 123)
	require.NoError(t, err)

	// getDataValue: retrieve the value
	val, err := internal.getDataValue(FooKey)
	require.NoError(t, err)
	require.Equal(t, 123, val)
}

func TestAssembly_putDataValue_NilKey(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	err := internal.putDataValue(nil, 1)
	require.Error(t, err)
}

func TestAssembly_putDataValue_DuplicateKey(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)

	// First putDataValue: should succeed
	err := internal.putDataValue(FooKey, 1)
	require.NoError(t, err)

	// Second putDataValue: should fail (duplicate key)
	err = internal.putDataValue(FooKey, 2)
	require.Error(t, err)

	// getDataValue: should return the original value (1)
	val, err := internal.getDataValue(FooKey)
	require.NoError(t, err)
	require.Equal(t, 1, val)
}

func TestAssembly_getDataValue_NilKey(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	_, err := internal.getDataValue(nil)
	require.Error(t, err)
}

func TestAssembly_getDataValue_MissingKey(t *testing.T) {
	asm, _ := NewAssembly()
	internal := asm.(*assembly)
	_, err := internal.getDataValue(FooKey)
	require.Error(t, err)
}

func TestAssembly_getData_BeforeBuild(t *testing.T) {
	asm, _ := NewAssembly()
	_, err := asm.getData(FooKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "getData: can only be called after Build has completed successfully")
}

func TestAssembly_getData_AfterBuild(t *testing.T) {
	// Create a module that produces FooKey
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			return b.putData(FooKey, 42)
		},
	}
	asm, err := NewAssembly(m1)
	require.NoError(t, err)

	// Before Build, getData should fail
	_, err = asm.getData(FooKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "getData: can only be called after Build has completed successfully")

	// Build the assembly
	err = asm.Build()
	require.NoError(t, err)

	// After Build, getData should succeed
	val, err := asm.getData(FooKey)
	require.NoError(t, err)
	require.Equal(t, 42, val)
}

func TestAssembly_getData_AfterBuildFailure(t *testing.T) {
	// Create a module that consumes a non-existent key
	m1 := &MockModule{
		NameValue:     "m1",
		ConsumesValue: Keys(FooKey), // No module produces this
	}
	asm, err := NewAssembly(m1)
	require.NoError(t, err)

	// Build should fail
	err = asm.Build()
	require.Error(t, err)

	// Even after Build failure, getData should still fail (not succeed)
	_, err = asm.getData(FooKey)
	require.Error(t, err)
	require.Contains(t, err.Error(), "getData: can only be called after Build has completed successfully")
}

func TestAssembly_DataGet_AfterBuild(t *testing.T) {
	// Create a module that produces FooKey
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(FooKey),
		ConfigureFunc: func(b Binder) error {
			return b.putData(FooKey, 42)
		},
	}
	asm, err := NewAssembly(m1)
	require.NoError(t, err)

	// Before Build, Data.Get should fail
	_, err = FooKey.Get(asm)
	require.Error(t, err)
	require.Contains(t, err.Error(), "getData: can only be called after Build has completed successfully")

	// Build the assembly
	err = asm.Build()
	require.NoError(t, err)

	// After Build, Data.Get should succeed
	val, err := FooKey.Get(asm)
	require.NoError(t, err)
	require.Equal(t, 42, val)
}

func TestAssembly_DuplicateProducers(t *testing.T) {
	// Create two modules that both declare they produce the same data key (FooKey)
	module1 := &MockModule{
		NameValue:     "module1",
		ProducesValue: Keys(FooKey),
		ConsumesValue: Keys(),
	}

	module2 := &MockModule{
		NameValue:     "module2",
		ProducesValue: Keys(FooKey),
		ConsumesValue: Keys(),
	}

	// Try to create an assembly with both modules
	_, err := NewAssembly(module1, module2)

	// Should get an error about duplicate producers
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate producer for data key")
	require.Contains(t, err.Error(), "foo")
	require.Contains(t, err.Error(), "both declare they produce it")
}

func TestAssembly_Install_RegistryValidationError_Produces(t *testing.T) {
	// Create a module that produces a key with a signature clash
	m1 := &MockModule{
		NameValue:     "m1",
		ProducesValue: Keys(ClashTestKey1),
	}

	m2 := &MockModule{
		NameValue:     "m2",
		ProducesValue: Keys(ClashTestKey2), // Same signature as ClashTestKey1
	}

	// Try to create an assembly with both modules
	// This should fail during installation due to registry validation error
	_, err := NewAssembly(m1, m2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "data key signature clash")
}

func TestAssembly_Install_RegistryValidationError_Consumes(t *testing.T) {
	// Create a module that consumes a key with a signature clash
	m1 := &MockModule{
		NameValue:     "m1",
		ConsumesValue: Keys(ClashTestKey1),
	}

	m2 := &MockModule{
		NameValue:     "m2",
		ConsumesValue: Keys(ClashTestKey2), // Same signature as ClashTestKey1
	}

	// Try to create an assembly with both modules
	// This should fail during installation due to registry validation error
	_, err := NewAssembly(m1, m2)
	require.Error(t, err)
	require.Contains(t, err.Error(), "data key signature clash")
}

func TestAssembly_Singleton_DuplicateInstallation(t *testing.T) {
	// Test that singleton modules can be installed multiple times without error
	singleton1 := &MockSingletonModule{NameValue: "singleton"}
	singleton2 := &MockSingletonModule{NameValue: "singleton"}

	assembly, err := NewAssembly(singleton1)
	require.NoError(t, err)

	// Installing the same singleton module again should not error
	err = assembly.Build()
	require.NoError(t, err)

	// Try to install the same singleton module through a binder
	// This would normally fail for non-singleton modules
	moduleWithBinder := &MockModule{
		NameValue:     "test",
		ProducesValue: Keys(),
		ConsumesValue: Keys(),
		ConfigureFunc: func(b Binder) error {
			// This should succeed for singleton modules
			return b.Install(singleton2)
		},
	}

	assembly2, err := NewAssembly(moduleWithBinder)
	require.NoError(t, err)

	err = assembly2.Build()
	require.NoError(t, err)
}

func TestAssembly_NonSingleton_DuplicateInstallation(t *testing.T) {
	// Test that non-singleton modules still error on duplicate installation
	module1 := &MockModule{NameValue: "non-singleton"}
	module2 := &MockModule{NameValue: "non-singleton"}

	// Installing the same non-singleton module again should error
	moduleWithBinder := &MockModule{
		NameValue:     "test",
		ProducesValue: Keys(),
		ConsumesValue: Keys(),
		ConfigureFunc: func(b Binder) error {
			// This should fail for non-singleton modules
			return b.Install(module2)
		},
	}

	assembly, err := NewAssembly(module1, moduleWithBinder)
	require.NoError(t, err)

	err = assembly.Build()
	require.Error(t, err)
	require.Contains(t, err.Error(), "module 'github.com/goosz/modz:non-singleton': already added")
}

func TestAssembly_Singleton_ModuleSignatureEquality(t *testing.T) {
	// Test that module signatures work correctly for map keys
	singleton1 := &MockSingletonModule{NameValue: "test"}
	singleton2 := &MockSingletonModule{NameValue: "test"}
	nonSingleton := &MockModule{NameValue: "test"}

	sig1 := newModuleSignature(singleton1)
	sig2 := newModuleSignature(singleton2)
	sig3 := newModuleSignature(nonSingleton)

	// Same name, both singletons - should be equal
	require.Equal(t, sig1, sig2)

	// Same name, different singleton status - should be equal (same package + name)
	require.Equal(t, sig1, sig3)
	require.Equal(t, sig2, sig3)
}

func TestAssembly_ModuleIdentity_TypeAndName(t *testing.T) {
	// Test that different types of modules can have the same name
	module1 := &MockModule{NameValue: "database"}
	module2 := &MockSingletonModule{NameValue: "database"}

	sig1 := newModuleSignature(module1)
	sig2 := newModuleSignature(module2)

	// Different types with same name in same package should be the same signature
	require.Equal(t, sig1, sig2)
	require.Equal(t, "github.com/goosz/modz:database", sig1.String())
	require.Equal(t, "github.com/goosz/modz:database", sig2.String())

	// Since they're the same signature, but one is a singleton, both should be installable
	assembly, err := NewAssembly(module1, module2)
	require.NoError(t, err)
	err = assembly.Build()
	require.NoError(t, err)
}
