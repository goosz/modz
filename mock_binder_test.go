package modz

import (
	"errors"
)

// MockBinder is a minimal implementation of Binder for unit tests.
type MockBinder struct {
	Store map[DataKey]any
}

func NewMockBinder() *MockBinder {
	return &MockBinder{Store: make(map[DataKey]any)}
}

func (m *MockBinder) Install(module Module) error {
	return errors.New("MockBinder.Install is not implemented")
}

func (m *MockBinder) getData(key DataKey) (any, error) {
	val, ok := m.Store[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return val, nil
}

func (m *MockBinder) putData(key DataKey, value any) error {
	if _, exists := m.Store[key]; exists {
		return errors.New("already set")
	}
	m.Store[key] = value
	return nil
}
