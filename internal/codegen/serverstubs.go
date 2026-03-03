package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ServerFunction represents a discovered server function
type ServerFunction struct {
	Name     string
	Package  string
	Params   []FuncParam
	Returns  []FuncReturn
	FilePath string
}

// FuncParam represents a function parameter
type FuncParam struct {
	Name string
	Type string
}

// FuncReturn represents a function return value
type FuncReturn struct {
	Type    string
	IsError bool
}

// DiscoverServerFunctions scans a directory for exported Go functions.
// It walks the directory looking for .go files, parses each with go/parser,
// finds exported functions (capitalized names), and extracts parameter
// names/types and return types. It skips init(), main(), test functions,
// and _test.go files.
func DiscoverServerFunctions(serverDir string) ([]ServerFunction, error) {
	fset := token.NewFileSet()

	// Filter out _test.go files
	filter := func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}

	packages, err := parser.ParseDir(fset, serverDir, filter, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server directory: %w", err)
	}

	var functions []ServerFunction

	for packageName, pkg := range packages {
		for fileName, file := range pkg.Files {
			fileFuncs := extractFunctions(packageName, fileName, file)
			functions = append(functions, fileFuncs...)
		}
	}

	// Sort for deterministic output
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})

	return functions, nil
}

// extractFunctions extracts exported function declarations from a parsed Go file.
// It skips unexported functions, init(), main(), and methods (functions with receivers).
func extractFunctions(packageName string, fileName string, file *ast.File) []ServerFunction {
	var functions []ServerFunction

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		// Skip methods (have a receiver)
		if fn.Recv != nil {
			continue
		}

		// Skip unexported functions
		if !fn.Name.IsExported() {
			continue
		}

		// Skip init and main
		name := fn.Name.Name
		if name == "init" || name == "main" {
			continue
		}

		sf := ServerFunction{
			Name:     name,
			Package:  packageName,
			FilePath: fileName,
			Params:   extractParams(fn.Type.Params),
			Returns:  extractReturns(fn.Type.Results),
		}

		functions = append(functions, sf)
	}

	return functions
}

// extractParams extracts parameter names and types from a field list
func extractParams(fields *ast.FieldList) []FuncParam {
	if fields == nil {
		return nil
	}

	var params []FuncParam

	for _, field := range fields.List {
		typeStr := exprToString(field.Type)

		if len(field.Names) == 0 {
			// Unnamed parameter
			params = append(params, FuncParam{
				Name: "",
				Type: typeStr,
			})
		} else {
			// Each name shares the same type
			for _, name := range field.Names {
				params = append(params, FuncParam{
					Name: name.Name,
					Type: typeStr,
				})
			}
		}
	}

	return params
}

// extractReturns extracts return types from a field list
func extractReturns(fields *ast.FieldList) []FuncReturn {
	if fields == nil {
		return nil
	}

	var returns []FuncReturn

	for _, field := range fields.List {
		typeStr := exprToString(field.Type)
		isError := typeStr == "error"

		if len(field.Names) == 0 {
			returns = append(returns, FuncReturn{
				Type:    typeStr,
				IsError: isError,
			})
		} else {
			for range field.Names {
				returns = append(returns, FuncReturn{
					Type:    typeStr,
					IsError: isError,
				})
			}
		}
	}

	return returns
}

// exprToString converts an AST expression to its string representation
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			// Slice
			return "[]" + exprToString(t.Elt)
		}
		// Array — use a basic representation
		return "[" + exprToString(t.Len) + "]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	case *ast.BasicLit:
		return t.Value
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + exprToString(t.Value)
		case ast.RECV:
			return "<-chan " + exprToString(t.Value)
		default:
			return "chan " + exprToString(t.Value)
		}
	case *ast.FuncType:
		return "func()"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// GenerateClientStubs generates WASM-side client stubs for server functions.
// The generated code has a //go:build js && wasm build tag and calls
// golem.CallServerFunction to invoke server functions via HTTP/JSON.
func GenerateClientStubs(functions []ServerFunction, moduleName string) (string, error) {
	var buf bytes.Buffer

	// Write build tag
	buf.WriteString("//go:build js && wasm\n\n")

	// Write package declaration
	buf.WriteString("package serverstubs\n\n")

	// Collect imports
	needsJSON := false
	for _, fn := range functions {
		if hasNonErrorReturn(fn.Returns) {
			needsJSON = true
			break
		}
	}

	// Write imports
	buf.WriteString("import (\n")
	if needsJSON {
		buf.WriteString("\t\"encoding/json\"\n")
	}
	buf.WriteString(fmt.Sprintf("\tgolem %q\n", filepath.Join(moduleName, "pkg/golem")))
	buf.WriteString(")\n\n")

	// Generate stub for each function
	for _, fn := range functions {
		stub := generateFunctionStub(fn)
		buf.WriteString(stub)
		buf.WriteString("\n")
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.String(), fmt.Errorf("failed to format generated code: %w", err)
	}

	return string(formatted), nil
}

// hasNonErrorReturn checks if a function has any non-error return values
func hasNonErrorReturn(returns []FuncReturn) bool {
	for _, r := range returns {
		if !r.IsError {
			return true
		}
	}
	return false
}

// generateFunctionStub generates a single client stub function
func generateFunctionStub(fn ServerFunction) string {
	var buf bytes.Buffer

	// Build parameter list
	paramList := buildParamList(fn.Params)

	// Build return list
	returnList := buildReturnList(fn.Returns)

	// Function signature
	buf.WriteString(fmt.Sprintf("func %s(%s) %s {\n", fn.Name, paramList, returnList))

	// Build args map
	if len(fn.Params) > 0 {
		buf.WriteString("\targs := map[string]interface{}{\n")
		for _, p := range fn.Params {
			name := p.Name
			if name == "" {
				name = "arg"
			}
			buf.WriteString(fmt.Sprintf("\t\t%q: %s,\n", name, name))
		}
		buf.WriteString("\t}\n")
	} else {
		buf.WriteString("\targs := map[string]interface{}{}\n")
	}

	// Call server function
	buf.WriteString(fmt.Sprintf("\tresult, err := golem.CallServerFunction(%q, args)\n", fn.Name))

	// Handle returns
	hasError := hasErrorReturn(fn.Returns)
	nonErrorReturns := getNonErrorReturns(fn.Returns)

	if len(nonErrorReturns) == 0 && hasError {
		// Only error return
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString("\t\treturn err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\t_ = result\n")
		buf.WriteString("\treturn nil\n")
	} else if len(nonErrorReturns) == 0 && !hasError {
		// No returns at all
		buf.WriteString("\t_ = result\n")
		buf.WriteString("\t_ = err\n")
	} else if len(nonErrorReturns) > 0 && hasError {
		// Has value return(s) and error
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString(fmt.Sprintf("\t\tvar zero %s\n", nonErrorReturns[0].Type))
		buf.WriteString("\t\treturn zero, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString(fmt.Sprintf("\tvar ret %s\n", nonErrorReturns[0].Type))
		buf.WriteString("\tif err := json.Unmarshal(result, &ret); err != nil {\n")
		buf.WriteString(fmt.Sprintf("\t\tvar zero %s\n", nonErrorReturns[0].Type))
		buf.WriteString("\t\treturn zero, err\n")
		buf.WriteString("\t}\n")
		buf.WriteString("\treturn ret, nil\n")
	} else {
		// Has value return(s) but no error
		buf.WriteString("\tif err != nil {\n")
		buf.WriteString(fmt.Sprintf("\t\tvar zero %s\n", nonErrorReturns[0].Type))
		buf.WriteString("\t\treturn zero\n")
		buf.WriteString("\t}\n")
		buf.WriteString(fmt.Sprintf("\tvar ret %s\n", nonErrorReturns[0].Type))
		buf.WriteString("\t_ = json.Unmarshal(result, &ret)\n")
		buf.WriteString("\treturn ret\n")
	}

	buf.WriteString("}\n")

	return buf.String()
}

// buildParamList builds the parameter list string for a function signature
func buildParamList(params []FuncParam) string {
	var parts []string
	for _, p := range params {
		if p.Name != "" {
			parts = append(parts, fmt.Sprintf("%s %s", p.Name, p.Type))
		} else {
			parts = append(parts, p.Type)
		}
	}
	return strings.Join(parts, ", ")
}

// buildReturnList builds the return type list string for a function signature
func buildReturnList(returns []FuncReturn) string {
	if len(returns) == 0 {
		return ""
	}

	if len(returns) == 1 {
		return returns[0].Type
	}

	var types []string
	for _, r := range returns {
		types = append(types, r.Type)
	}
	return "(" + strings.Join(types, ", ") + ")"
}

// hasErrorReturn checks if any return value is an error
func hasErrorReturn(returns []FuncReturn) bool {
	for _, r := range returns {
		if r.IsError {
			return true
		}
	}
	return false
}

// getNonErrorReturns returns all non-error return values
func getNonErrorReturns(returns []FuncReturn) []FuncReturn {
	var result []FuncReturn
	for _, r := range returns {
		if !r.IsError {
			result = append(result, r)
		}
	}
	return result
}
