package modz

var (
	ProducedKey = NewData[string]("produced")
	ConsumedKey = NewData[int]("consumed")
	FooKey      = NewData[int]("foo")
	BarKey      = NewData[int]("bar")

	// Keys for registry validation testing
	ClashTestKey1 = NewData[int]("clash-test-1")
	ClashTestKey2 = NewData[int]("clash-test-1") // Same signature as ClashTestKey1
)

// MockModule is a minimal implementation of Module for unit tests.
type MockModule struct {
	NameValue     string
	ProducesValue DataKeys
	ConsumesValue DataKeys
	ConfigureFunc func(Binder) error
}

func (m *MockModule) Name() string       { return m.NameValue }
func (m *MockModule) Produces() DataKeys { return m.ProducesValue }
func (m *MockModule) Consumes() DataKeys { return m.ConsumesValue }
func (m *MockModule) Configure(binder Binder) error {
	if m.ConfigureFunc != nil {
		return m.ConfigureFunc(binder)
	}
	return nil
}

// MockSingletonModule is a mock module with singleton behavior for testing.
type MockSingletonModule struct {
	NameValue     string
	ProducesValue DataKeys
	ConsumesValue DataKeys
	ConfigureFunc func(Binder) error
	Singleton
}

func (m *MockSingletonModule) Name() string       { return m.NameValue }
func (m *MockSingletonModule) Produces() DataKeys { return m.ProducesValue }
func (m *MockSingletonModule) Consumes() DataKeys { return m.ConsumesValue }
func (m *MockSingletonModule) Configure(binder Binder) error {
	if m.ConfigureFunc != nil {
		return m.ConfigureFunc(binder)
	}
	return nil
}
