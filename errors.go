package modz

import "fmt"

// ConfigurationError represents an error that occurred during module configuration.
// It provides context about which module encountered the error and what operation failed.
type ConfigurationError struct {
	ModuleID  string
	Operation string
	Err       error
}

func (e *ConfigurationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("module '%s' %s: %v", e.ModuleID, e.Operation, e.Err)
	}
	return fmt.Sprintf("module '%s' %s failed", e.ModuleID, e.Operation)
}

// newPhaseError creates a consistent error for phase violations
func newPhaseError(operation string) error {
	return fmt.Errorf("%s: can only be called during configuration phase", operation)
}

// newUndeclaredKeyError creates a consistent error for undeclared key access
func newUndeclaredKeyError(moduleName string, key DataKey, declarationType string) error {
	return fmt.Errorf("module '%s' did not declare '%s' in %s", moduleName, key, declarationType)
}

// newInstallError creates a consistent error for module installation failures
func newInstallError(moduleName string, message string) error {
	return fmt.Errorf("module '%s': %s", moduleName, message)
}

// newDataOperationError creates a consistent error for data operation failures
func newDataOperationError(key DataKey, message string) error {
	return fmt.Errorf("data key '%s': %s", key, message)
}

// newFailFastError creates a consistent error for fail-fast behavior
func newFailFastError(operation string, previousError error) error {
	return fmt.Errorf("%s: failed due to previous error: %w", operation, previousError)
}
