package serverfuncs

import (
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadAndCallGoPlugin_FileNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	err := LoadAndCallGoPlugin("../../testdata", "nonexistent", rec, httptest.NewRequest("GET", "/", nil))
	if err == nil || err.Error() == "" {
		t.Error("expected error for missing plugin file")
	}
}

func TestDiscoverServerFunctions(t *testing.T) {
	dir := "../../testdata"
	os.MkdirAll(dir+"/server/go", 0755)
	os.MkdirAll(dir+"/server/ts", 0755)
	os.WriteFile(dir+"/server/go/hello.go", []byte("package main\nfunc Handler(){}"), 0644)
	os.WriteFile(dir+"/server/ts/hello.ts", []byte("export default () => {}"), 0644)
	funcs, err := DiscoverServerFunctions(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(funcs) < 2 {
		t.Errorf("expected at least 2 functions, got %d", len(funcs))
	}
	os.RemoveAll(dir)
}
