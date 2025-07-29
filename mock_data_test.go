package modz

import (
	"errors"
)

// MockDataReadWriter is a minimal implementation of DataReader and DataWriter for unit tests.
type MockDataReadWriter struct {
	Store map[DataKey]any
}

// Ensure that MockDataReadWriter implements the required interfaces.
var _ DataReader = (*MockDataReadWriter)(nil)
var _ DataWriter = (*MockDataReadWriter)(nil)

func NewMockDataReadWriter() *MockDataReadWriter {
	return &MockDataReadWriter{Store: make(map[DataKey]any)}
}

func (m *MockDataReadWriter) getData(key DataKey) (any, error) {
	val, ok := m.Store[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return val, nil
}

func (m *MockDataReadWriter) putData(key DataKey, value any) error {
	if _, exists := m.Store[key]; exists {
		return errors.New("already set")
	}
	m.Store[key] = value
	return nil
}
