package functions

import (
	"github.com/Nu11ified/golem/internal/functions"
)

// Register allows user packages to register their functions with the framework
func Register(serviceName, functionName string, fn interface{}) error {
	return functions.RegisterGlobalFunction(serviceName, functionName, fn)
}

// GetRegistry returns the current function registry for use by the framework
func GetRegistry() *functions.Registry {
	return functions.GetGlobalRegistry()
}

// HasFunctions returns true if any functions have been registered
func HasFunctions() bool {
	registry := functions.GetGlobalRegistry()
	return len(registry.ListFunctions("")) > 0
}
