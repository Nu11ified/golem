//go:build !js || !wasm

package codegen_test

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/Nu11ified/golem/internal/codegen"
)

func TestDiscoverServerFunctions_Simple(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"api.go": `package server

func GetPosts() []string {
	return nil
}
`,
	})

	funcs, err := codegen.DiscoverServerFunctions(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}

	if funcs[0].Name != "GetPosts" {
		t.Errorf("expected function name 'GetPosts', got '%s'", funcs[0].Name)
	}

	if len(funcs[0].Returns) != 1 {
		t.Fatalf("expected 1 return value, got %d", len(funcs[0].Returns))
	}

	if funcs[0].Returns[0].Type != "[]string" {
		t.Errorf("expected return type '[]string', got '%s'", funcs[0].Returns[0].Type)
	}
}

func TestDiscoverServerFunctions_WithParams(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"api.go": `package server

func GetPostsByCategory(category string, limit int) []string {
	return nil
}
`,
	})

	funcs, err := codegen.DiscoverServerFunctions(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}

	fn := funcs[0]
	if len(fn.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fn.Params))
	}

	if fn.Params[0].Name != "category" {
		t.Errorf("expected first param name 'category', got '%s'", fn.Params[0].Name)
	}
	if fn.Params[0].Type != "string" {
		t.Errorf("expected first param type 'string', got '%s'", fn.Params[0].Type)
	}

	if fn.Params[1].Name != "limit" {
		t.Errorf("expected second param name 'limit', got '%s'", fn.Params[1].Name)
	}
	if fn.Params[1].Type != "int" {
		t.Errorf("expected second param type 'int', got '%s'", fn.Params[1].Type)
	}
}

func TestDiscoverServerFunctions_WithErrorReturn(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"api.go": `package server

type Post struct {
	Title string
}

func GetPosts(category string) ([]Post, error) {
	return nil, nil
}
`,
	})

	funcs, err := codegen.DiscoverServerFunctions(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(funcs))
	}

	fn := funcs[0]
	if len(fn.Returns) != 2 {
		t.Fatalf("expected 2 return values, got %d", len(fn.Returns))
	}

	if fn.Returns[0].Type != "[]Post" {
		t.Errorf("expected first return type '[]Post', got '%s'", fn.Returns[0].Type)
	}
	if fn.Returns[0].IsError {
		t.Error("expected first return to not be error")
	}

	if fn.Returns[1].Type != "error" {
		t.Errorf("expected second return type 'error', got '%s'", fn.Returns[1].Type)
	}
	if !fn.Returns[1].IsError {
		t.Error("expected second return to be error")
	}
}

func TestDiscoverServerFunctions_SkipsUnexported(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"api.go": `package server

func getPosts() []string {
	return nil
}

func GetUsers() []string {
	return nil
}
`,
	})

	funcs, err := codegen.DiscoverServerFunctions(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function (only exported), got %d", len(funcs))
	}

	if funcs[0].Name != "GetUsers" {
		t.Errorf("expected function name 'GetUsers', got '%s'", funcs[0].Name)
	}
}

func TestDiscoverServerFunctions_SkipsTestFiles(t *testing.T) {
	dir := createTestApp(t, map[string]string{
		"api.go": `package server

func GetPosts() []string {
	return nil
}
`,
		"api_test.go": `package server

func TestHelper() string {
	return ""
}
`,
	})

	funcs, err := codegen.DiscoverServerFunctions(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(funcs) != 1 {
		t.Fatalf("expected 1 function (test file skipped), got %d", len(funcs))
	}

	if funcs[0].Name != "GetPosts" {
		t.Errorf("expected function name 'GetPosts', got '%s'", funcs[0].Name)
	}
}

func TestGenerateClientStubs_ValidGoSyntax(t *testing.T) {
	funcs := []codegen.ServerFunction{
		{
			Name:    "GetPosts",
			Package: "server",
			Params: []codegen.FuncParam{
				{Name: "category", Type: "string"},
			},
			Returns: []codegen.FuncReturn{
				{Type: "[]string", IsError: false},
				{Type: "error", IsError: true},
			},
		},
	}

	code, err := codegen.GenerateClientStubs(funcs, "myapp")
	if err != nil {
		t.Fatal(err)
	}

	// Verify it parses as valid Go
	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "stubs.go", code, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("generated code is not valid Go: %v\nCode:\n%s", parseErr, code)
	}
}

func TestGenerateClientStubs_ContainsFunction(t *testing.T) {
	funcs := []codegen.ServerFunction{
		{
			Name:    "GetPosts",
			Package: "server",
			Params: []codegen.FuncParam{
				{Name: "category", Type: "string"},
			},
			Returns: []codegen.FuncReturn{
				{Type: "[]string", IsError: false},
				{Type: "error", IsError: true},
			},
		},
	}

	code, err := codegen.GenerateClientStubs(funcs, "myapp")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(code, "func GetPosts(") {
		t.Errorf("generated code does not contain function signature\nCode:\n%s", code)
	}

	if !strings.Contains(code, "category string") {
		t.Errorf("generated code does not contain parameter\nCode:\n%s", code)
	}

	if !strings.Contains(code, "CallServerFunction") {
		t.Errorf("generated code does not contain CallServerFunction call\nCode:\n%s", code)
	}
}

func TestGenerateClientStubs_HasBuildTag(t *testing.T) {
	funcs := []codegen.ServerFunction{
		{
			Name:    "GetPosts",
			Package: "server",
			Params:  []codegen.FuncParam{},
			Returns: []codegen.FuncReturn{
				{Type: "[]string", IsError: false},
			},
		},
	}

	code, err := codegen.GenerateClientStubs(funcs, "myapp")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(code, "//go:build js && wasm") {
		t.Errorf("generated code does not contain build tag\nCode:\n%s", code)
	}
}
