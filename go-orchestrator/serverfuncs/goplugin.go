package serverfuncs

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

type cachedPlugin struct {
	plugin  *plugin.Plugin
	modTime int64
}

var (
	pluginCache = make(map[string]cachedPlugin)
	cacheMu     sync.Mutex
)

// LoadAndCallGoPlugin loads the plugin for the given function name and calls its Handler.
// It caches plugins in memory and reloads if the .so file has changed.
func LoadAndCallGoPlugin(baseDir, functionName string, w http.ResponseWriter, r *http.Request) error {
	pluginPath := filepath.Join(baseDir, "server/go", functionName+".so")
	stat, err := os.Stat(pluginPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", pluginPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat plugin: %w", err)
	}
	modTime := stat.ModTime().UnixNano()

	cacheMu.Lock()
	cached, found := pluginCache[pluginPath]
	cacheMu.Unlock()

	var p *plugin.Plugin
	if found && cached.modTime == modTime {
		p = cached.plugin
	} else {
		p, err = plugin.Open(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to open plugin: %w", err)
		}
		cacheMu.Lock()
		pluginCache[pluginPath] = cachedPlugin{plugin: p, modTime: modTime}
		cacheMu.Unlock()
	}

	sym, err := p.Lookup("Handler")
	if err != nil {
		return fmt.Errorf("failed to find Handler symbol: %w", err)
	}
	handler, ok := sym.(func(http.ResponseWriter, *http.Request))
	if !ok {
		return fmt.Errorf("Handler has wrong type signature")
	}
	defer func() {
		if rec := recover(); rec != nil {
			http.Error(w, fmt.Sprintf(`{"error": "panic in handler: %v"}`, rec), http.StatusInternalServerError)
		}
	}()
	handler(w, r)
	return nil
}
