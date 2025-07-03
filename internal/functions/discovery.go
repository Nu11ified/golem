package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"google.golang.org/protobuf/types/known/anypb"

	pb "github.com/Nu11ified/golem/proto/gen/proto"
)

// Global registry for user functions to self-register
var globalRegistry *Registry
var globalMutex sync.Mutex

func init() {
	globalRegistry = NewRegistry()
}

// RegisterGlobalFunction allows user packages to register their functions
func RegisterGlobalFunction(serviceName, functionName string, fn interface{}) error {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	return globalRegistry.RegisterFunction(serviceName, functionName, fn)
}

// GetGlobalRegistry returns the global registry with all registered functions
func GetGlobalRegistry() *Registry {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	// Create a new registry and copy all functions from global registry
	registry := NewRegistry()
	for key, meta := range globalRegistry.functions {
		registry.functions[key] = meta
	}

	return registry
}

// Registry holds all discovered server functions
type Registry struct {
	functions map[string]*FunctionMeta
	packages  map[string]interface{} // Package instances
	mutex     sync.RWMutex
}

// FunctionMeta contains metadata about a server function
type FunctionMeta struct {
	Name        string
	ServiceName string
	Package     string
	Function    reflect.Value
	Type        reflect.Type
	ArgTypes    []string
	ReturnType  string
	Description string
}

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	return &Registry{
		functions: make(map[string]*FunctionMeta),
		packages:  make(map[string]interface{}),
	}
}

// RegisterPackage registers all exported functions from a package
func (r *Registry) RegisterPackage(packageName string, pkg interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.packages[packageName] = pkg

	pkgValue := reflect.ValueOf(pkg)
	pkgType := reflect.TypeOf(pkg)

	// If it's a pointer to a struct, get the underlying type
	if pkgType.Kind() == reflect.Ptr {
		pkgValue = pkgValue.Elem()
		pkgType = pkgType.Elem()
	}

	// Register all exported methods
	for i := 0; i < pkgType.NumMethod(); i++ {
		method := pkgType.Method(i)
		if method.IsExported() {
			meta := &FunctionMeta{
				Name:        method.Name,
				ServiceName: packageName,
				Package:     packageName,
				Function:    pkgValue.MethodByName(method.Name),
				Type:        method.Type,
			}

			// Extract argument and return types
			meta.ArgTypes = r.extractArgTypes(method.Type)
			meta.ReturnType = r.extractReturnType(method.Type)

			key := fmt.Sprintf("%s.%s", packageName, method.Name)
			r.functions[key] = meta
		}
	}

	return nil
}

// RegisterFunction registers a single function
func (r *Registry) RegisterFunction(serviceName, functionName string, fn interface{}) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("provided value is not a function")
	}

	meta := &FunctionMeta{
		Name:        functionName,
		ServiceName: serviceName,
		Package:     serviceName,
		Function:    fnValue,
		Type:        fnType,
		ArgTypes:    r.extractArgTypes(fnType),
		ReturnType:  r.extractReturnType(fnType),
	}

	key := fmt.Sprintf("%s.%s", serviceName, functionName)
	r.functions[key] = meta

	return nil
}

// DiscoverFunctions automatically discovers functions from source files
func (r *Registry) DiscoverFunctions(serverDir string) error {
	// Parse Go files in the server directory
	fset := token.NewFileSet()

	packages, err := parser.ParseDir(fset, serverDir, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse server directory: %w", err)
	}

	for packageName, pkg := range packages {
		if packageName == "main" {
			continue // Skip main packages
		}

		for fileName, file := range pkg.Files {
			if strings.HasSuffix(fileName, "_test.go") {
				continue // Skip test files
			}

			r.parseFileForFunctions(packageName, file)
		}
	}

	return nil
}

// parseFileForFunctions parses a Go file and extracts function information
func (r *Registry) parseFileForFunctions(packageName string, file *ast.File) {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.IsExported() {
				// Extract function metadata from AST
				meta := &FunctionMeta{
					Name:        fn.Name.Name,
					ServiceName: packageName,
					Package:     packageName,
					Description: r.extractDocString(fn.Doc),
				}

				// For now, we'll register the metadata
				// The actual function values will be registered when packages are loaded
				key := fmt.Sprintf("%s.%s", packageName, fn.Name.Name)

				r.mutex.Lock()
				if _, exists := r.functions[key]; !exists {
					r.functions[key] = meta
				}
				r.mutex.Unlock()
			}
		}
	}
}

// extractDocString extracts documentation from comments
func (r *Registry) extractDocString(commentGroup *ast.CommentGroup) string {
	if commentGroup == nil {
		return ""
	}

	var doc strings.Builder
	for _, comment := range commentGroup.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if text != "" {
			doc.WriteString(text)
			doc.WriteString(" ")
		}
	}

	return strings.TrimSpace(doc.String())
}

// extractArgTypes extracts argument types from function type
func (r *Registry) extractArgTypes(fnType reflect.Type) []string {
	var argTypes []string

	for i := 0; i < fnType.NumIn(); i++ {
		argType := fnType.In(i)
		// Skip context.Context as first parameter
		if i == 0 && argType.String() == "context.Context" {
			continue
		}
		argTypes = append(argTypes, argType.String())
	}

	return argTypes
}

// extractReturnType extracts return type from function type
func (r *Registry) extractReturnType(fnType reflect.Type) string {
	switch fnType.NumOut() {
	case 0:
		return "void"
	case 1:
		return fnType.Out(0).String()
	case 2:
		// Assume (result, error) pattern
		if fnType.Out(1).String() == "error" {
			return fnType.Out(0).String()
		}
		return fmt.Sprintf("(%s, %s)", fnType.Out(0).String(), fnType.Out(1).String())
	default:
		var types []string
		for i := 0; i < fnType.NumOut(); i++ {
			types = append(types, fnType.Out(i).String())
		}
		return fmt.Sprintf("(%s)", strings.Join(types, ", "))
	}
}

// CallFunction calls a registered function with the given arguments
func (r *Registry) CallFunction(ctx context.Context, serviceName, functionName string, args []*anypb.Any) (*anypb.Any, error) {
	r.mutex.RLock()
	key := fmt.Sprintf("%s.%s", serviceName, functionName)
	meta, exists := r.functions[key]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("function %s not found", key)
	}

	if !meta.Function.IsValid() {
		return nil, fmt.Errorf("function %s not properly registered", key)
	}

	// Convert protobuf Any arguments to Go values
	callArgs, err := r.convertArgs(ctx, meta.Type, args)
	if err != nil {
		return nil, fmt.Errorf("failed to convert arguments: %w", err)
	}

	// Call the function
	results := meta.Function.Call(callArgs)

	// Handle function results
	return r.convertResult(results)
}

// convertArgs converts protobuf Any arguments to Go reflect.Values
func (r *Registry) convertArgs(ctx context.Context, fnType reflect.Type, args []*anypb.Any) ([]reflect.Value, error) {
	var callArgs []reflect.Value

	// Check if first parameter is context.Context
	startIndex := 0
	if fnType.NumIn() > 0 && fnType.In(0).String() == "context.Context" {
		callArgs = append(callArgs, reflect.ValueOf(ctx))
		startIndex = 1
	}

	// Convert remaining arguments
	for i, arg := range args {
		paramIndex := startIndex + i
		if paramIndex >= fnType.NumIn() {
			return nil, fmt.Errorf("too many arguments provided")
		}

		paramType := fnType.In(paramIndex)
		value, err := r.convertAnyToValue(arg, paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert argument %d: %w", i, err)
		}

		callArgs = append(callArgs, value)
	}

	// Check if we have enough arguments
	requiredArgs := fnType.NumIn()
	if fnType.NumIn() > 0 && fnType.In(0).String() == "context.Context" {
		requiredArgs--
	}

	if len(args) != requiredArgs {
		return nil, fmt.Errorf("expected %d arguments, got %d", requiredArgs, len(args))
	}

	return callArgs, nil
}

// convertAnyToValue converts a protobuf Any to a Go reflect.Value
func (r *Registry) convertAnyToValue(any *anypb.Any, targetType reflect.Type) (reflect.Value, error) {
	// Extract JSON data from the Any message
	jsonData := any.GetValue()

	// Create a new instance of the target type
	value := reflect.New(targetType).Interface()

	// Unmarshal JSON to the target type
	if err := json.Unmarshal(jsonData, value); err != nil {
		// Try direct type conversion for primitive types
		return r.convertPrimitive(jsonData, targetType)
	}

	return reflect.ValueOf(value).Elem(), nil
}

// convertPrimitive handles primitive type conversions
func (r *Registry) convertPrimitive(jsonData []byte, targetType reflect.Type) (reflect.Value, error) {
	var rawValue interface{}
	if err := json.Unmarshal(jsonData, &rawValue); err != nil {
		return reflect.Value{}, err
	}

	// Convert based on target type
	switch targetType.Kind() {
	case reflect.String:
		if str, ok := rawValue.(string); ok {
			return reflect.ValueOf(str), nil
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		if num, ok := rawValue.(float64); ok {
			return reflect.ValueOf(int64(num)).Convert(targetType), nil
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := rawValue.(float64); ok {
			return reflect.ValueOf(num).Convert(targetType), nil
		}
	case reflect.Bool:
		if b, ok := rawValue.(bool); ok {
			return reflect.ValueOf(b), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %s", rawValue, targetType.String())
}

// convertResult converts function results to protobuf Any
func (r *Registry) convertResult(results []reflect.Value) (*anypb.Any, error) {
	switch len(results) {
	case 0:
		// No return value
		return anypb.New(&pb.FunctionResponse{Success: true})
	case 1:
		// Single return value
		result := results[0].Interface()
		return r.valueToAny(result)
	case 2:
		// Assume (result, error) pattern
		result := results[0].Interface()
		errVal := results[1]

		if !errVal.IsNil() {
			err := errVal.Interface().(error)
			return nil, err
		}

		return r.valueToAny(result)
	default:
		// Multiple return values - wrap in slice
		var resultSlice []interface{}
		for _, result := range results {
			resultSlice = append(resultSlice, result.Interface())
		}
		return r.valueToAny(resultSlice)
	}
}

// valueToAny converts a Go value to protobuf Any
func (r *Registry) valueToAny(value interface{}) (*anypb.Any, error) {
	// Serialize to JSON first
	jsonData, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	// Create Any message with JSON data
	return &anypb.Any{
		TypeUrl: "type.googleapis.com/google.protobuf.Value",
		Value:   jsonData,
	}, nil
}

// ListFunctions returns all registered functions
func (r *Registry) ListFunctions(serviceName string) []*pb.FunctionInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var functions []*pb.FunctionInfo

	for _, meta := range r.functions {
		// Filter by service name if specified
		if serviceName != "" && meta.ServiceName != serviceName {
			continue
		}

		functions = append(functions, &pb.FunctionInfo{
			Name:        meta.Name,
			ServiceName: meta.ServiceName,
			ArgTypes:    meta.ArgTypes,
			ReturnType:  meta.ReturnType,
			Description: meta.Description,
		})
	}

	return functions
}

// GetFunction returns metadata for a specific function
func (r *Registry) GetFunction(serviceName, functionName string) (*FunctionMeta, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	key := fmt.Sprintf("%s.%s", serviceName, functionName)
	meta, exists := r.functions[key]
	return meta, exists
}

// BuildAndImportServerPackages dynamically builds and imports server packages
// to trigger their init() functions and register their functions
func (r *Registry) BuildAndImportServerPackages(serverDir string) error {
	// Find all .go files in the server directory
	goFiles, err := r.findGoFiles(serverDir)
	if err != nil {
		return fmt.Errorf("failed to find Go files in %s: %w", serverDir, err)
	}

	if len(goFiles) == 0 {
		log.Printf("No Go files found in %s", serverDir)
		return nil
	}

	// Generate import statements and create a temporary file to trigger init() functions
	return r.generateAndBuildImports(serverDir, goFiles)
}

// findGoFiles recursively finds all .go files in a directory
func (r *Registry) findGoFiles(dir string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	return goFiles, err
}

// generateAndBuildImports creates import statements for server packages
func (r *Registry) generateAndBuildImports(serverDir string, goFiles []string) error {
	// Parse the go.mod file to get the module name
	moduleName, err := r.getModuleName()
	if err != nil {
		log.Printf("Warning: Could not determine module name: %v", err)
		return nil
	}

	// Group files by package
	packages := make(map[string]bool)
	for _, file := range goFiles {
		// Get the directory of the file relative to the current directory
		dir := filepath.Dir(file)
		if dir != "." {
			// Convert file path to import path
			importPath := filepath.Join(moduleName, dir)
			packages[importPath] = true
		}
	}

	// Create temporary import file
	if len(packages) > 0 {
		return r.createImportFile(packages)
	}

	return nil
}

// GetModuleName reads the module name from go.mod
func GetModuleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// getModuleName reads the module name from go.mod (private method)
func (r *Registry) getModuleName() (string, error) {
	return GetModuleName()
}

// createImportFile creates a temporary file that imports all server packages
func (r *Registry) createImportFile(packages map[string]bool) error {
	// Create .golem directory if it doesn't exist
	if err := os.MkdirAll(".golem", 0755); err != nil {
		return err
	}

	// Generate import file content
	var imports []string
	for pkg := range packages {
		imports = append(imports, fmt.Sprintf(`_ "%s"`, pkg))
	}

	content := fmt.Sprintf(`// Auto-generated file to import server packages
// This file triggers init() functions in server packages
package main

import (
%s
)

func init() {
	// Server packages imported above will have their init() functions called
	// which should register their functions with the global registry
}
`, strings.Join(imports, "\n\t"))

	// Write the import file
	importFile := ".golem/server_imports.go"
	if err := os.WriteFile(importFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write import file: %w", err)
	}

	log.Printf("Generated server import file with %d packages", len(packages))
	return nil
}

// RegisterFromGlobal copies all functions from the global registry to this registry
func (r *Registry) RegisterFromGlobal() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	globalMutex.Lock()
	defer globalMutex.Unlock()

	// Copy all functions from global registry
	for key, meta := range globalRegistry.functions {
		r.functions[key] = meta
	}

	log.Printf("Copied %d functions from global registry", len(globalRegistry.functions))
	return nil
}
