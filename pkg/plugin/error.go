package plugin

import "fmt"

// PluginError represents a plugin-specific error
type PluginError struct {
	// Plugin name
	Plugin string
	
	// Plugin type (parser, renderer, theme)
	Type string
	
	// Operation that failed
	Operation string
	
	// Underlying error
	Err error
}

// Error implements the error interface
func (e *PluginError) Error() string {
	return fmt.Sprintf("plugin error [%s/%s]: %s failed: %v", e.Type, e.Plugin, e.Operation, e.Err)
}

// Unwrap returns the underlying error
func (e *PluginError) Unwrap() error {
	return e.Err
}

// NewPluginError creates a new plugin error
func NewPluginError(pluginType, pluginName, operation string, err error) *PluginError {
	return &PluginError{
		Plugin:    pluginName,
		Type:      pluginType,
		Operation: operation,
		Err:       err,
	}
}

// RegistrationError represents a plugin registration error
type RegistrationError struct {
	// Plugin name
	Plugin string
	
	// Plugin type
	Type string
	
	// Reason for failure
	Reason string
}

// Error implements the error interface
func (e *RegistrationError) Error() string {
	return fmt.Sprintf("plugin registration error [%s/%s]: %s", e.Type, e.Plugin, e.Reason)
}

// NewRegistrationError creates a new registration error
func NewRegistrationError(pluginType, pluginName, reason string) *RegistrationError {
	return &RegistrationError{
		Plugin: pluginName,
		Type:   pluginType,
		Reason: reason,
	}
}

// ConfigurationError represents a plugin configuration error
type ConfigurationError struct {
	// Plugin name
	Plugin string
	
	// Plugin type
	Type string
	
	// Configuration key that failed
	Key string
	
	// Underlying error
	Err error
}

// Error implements the error interface
func (e *ConfigurationError) Error() string {
	return fmt.Sprintf("plugin configuration error [%s/%s]: key '%s' failed: %v", e.Type, e.Plugin, e.Key, e.Err)
}

// Unwrap returns the underlying error
func (e *ConfigurationError) Unwrap() error {
	return e.Err
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(pluginType, pluginName, key string, err error) *ConfigurationError {
	return &ConfigurationError{
		Plugin: pluginName,
		Type:   pluginType,
		Key:    key,
		Err:    err,
	}
}

// SafeCall safely calls a plugin method with error handling
func SafeCall(pluginType, pluginName, operation string, fn func() error) error {
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error
			err := fmt.Errorf("panic in plugin: %v", r)
			// Log the panic (in a real implementation, you'd use a proper logger)
			fmt.Printf("Plugin panic recovered: %v\n", NewPluginError(pluginType, pluginName, operation, err))
		}
	}()
	
	if err := fn(); err != nil {
		return NewPluginError(pluginType, pluginName, operation, err)
	}
	
	return nil
}

// SafeCallWithResult safely calls a plugin method that returns a result
func SafeCallWithResult[T any](pluginType, pluginName, operation string, fn func() (T, error)) (T, error) {
	var zero T
	
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error
			err := fmt.Errorf("panic in plugin: %v", r)
			// Log the panic (in a real implementation, you'd use a proper logger)
			fmt.Printf("Plugin panic recovered: %v\n", NewPluginError(pluginType, pluginName, operation, err))
		}
	}()
	
	result, err := fn()
	if err != nil {
		return zero, NewPluginError(pluginType, pluginName, operation, err)
	}
	
	return result, nil
}

// IsPluginError checks if an error is a plugin error
func IsPluginError(err error) bool {
	_, ok := err.(*PluginError)
	return ok
}

// IsRegistrationError checks if an error is a registration error
func IsRegistrationError(err error) bool {
	_, ok := err.(*RegistrationError)
	return ok
}

// IsConfigurationError checks if an error is a configuration error
func IsConfigurationError(err error) bool {
	_, ok := err.(*ConfigurationError)
	return ok
}

// GetPluginFromError extracts plugin information from an error
func GetPluginFromError(err error) (pluginType, pluginName string, ok bool) {
	if pluginErr, ok := err.(*PluginError); ok {
		return pluginErr.Type, pluginErr.Plugin, true
	}
	
	if regErr, ok := err.(*RegistrationError); ok {
		return regErr.Type, regErr.Plugin, true
	}
	
	if configErr, ok := err.(*ConfigurationError); ok {
		return configErr.Type, configErr.Plugin, true
	}
	
	return "", "", false
}