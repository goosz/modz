package modz

var (
	ProducedKey = NewData[string]("produced")
	ConsumedKey = NewData[int]("consumed")
	FooKey      = NewData[int]("foo")
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
