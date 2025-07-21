package modz

// MockModule is a minimal implementation of Module for unit tests.
type MockModule struct {
	NameValue string
}

func (m *MockModule) Name() string           { return m.NameValue }
func (m *MockModule) Configure(Binder) error { return nil }
func (m *MockModule) Consumes() DataKeys     { return nil }
func (m *MockModule) Produces() DataKeys     { return nil }
