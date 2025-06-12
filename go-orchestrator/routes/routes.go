package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lpar/gzipped/v2"

	"go-orchestrator/schema"
	"go-orchestrator/serverfuncs"
)

var baseDir = getBaseDir()

func getBaseDir() string {
	if dir := os.Getenv("BASE_DIR"); dir != "" {
		return dir
	}
	// Default: use "../" for dev, "./" for Docker/prod
	if _, err := os.Stat("../user-app/pages"); err == nil {
		return "../"
	}
	return "./"
}

var nodeRendererURL = getNodeRendererURL()

func getNodeRendererURL() string {
	host := os.Getenv("NODE_RENDERER_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("NODE_PORT")
	if port == "" {
		port = "3001"
	}
	return fmt.Sprintf("http://%s:%s/render", host, port)
}

type RenderPayload struct {
	ComponentPath string                 `json:"componentPath"`
	LayoutPath    string                 `json:"layoutPath,omitempty"`
	Props         map[string]interface{} `json:"props"`
}

type RouteInfo struct {
	PagePath   string
	LayoutPath string
}

var routeMap map[string]RouteInfo

func SetupRouter() http.Handler {
	routeMap = buildRouteMap(baseDir + "user-app/pages")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// Add gziphandler middleware for dynamic responses
	r.Use(gziphandler.GzipHandler)

	// Serve the UI schema JSON
	r.Get("/ui-schema", func(w http.ResponseWriter, r *http.Request) {
		schemaJSON, err := schema.GetUISchemaJSON()
		if err != nil {
			http.Error(w, "Failed to load UI schema", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/schema+json")
		w.Write(schemaJSON)
	})

	// Serve a sample UI JSON
	r.Get("/ui", func(w http.ResponseWriter, r *http.Request) {
		ui, err := schema.GetSampleUI()
		if err != nil {
			http.Error(w, "Failed to generate sample UI", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ui)
	})

	// Serve static files (client.js, client.js.map) with precompressed support
	r.Handle("/client.js", gzipped.FileServer(gzipped.Dir(baseDir+"node-renderer/dist")))
	r.Handle("/client.js.map", gzipped.FileServer(gzipped.Dir(baseDir+"node-renderer/dist")))

	// Handle favicon.ico requests with 404
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Handle .well-known and other static paths with 404
	r.Handle("/.well-known/*", http.HandlerFunc(http.NotFound))

	// Serve static assets from user-app/public if they exist
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		publicPath := baseDir + "user-app/public" + r.URL.Path
		if fileExists(publicPath) {
			http.ServeFile(w, r, publicPath)
			return
		}
		// Only SSR if the route exists in routeMap
		if info, ok := routeMap[r.URL.Path]; ok {
			handlePageRenderWithLayout(w, r, info)
			return
		}
		http.NotFound(w, r)
	})

	// API route for Go server functions
	r.Post("/api/go/{functionName}", handleGoFunction)
	// API route for TypeScript server functions
	r.Post("/api/ts/{functionName}", handleTSFunction)

	// New API route for listing all available functions
	r.Get("/api/functions", handleFunctionList)

	return r
}

func buildRouteMap(pagesDir string) map[string]RouteInfo {
	routes := make(map[string]RouteInfo)
	_ = filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".tsx" && !strings.HasSuffix(path, "layout.tsx") {
			rel, _ := filepath.Rel(pagesDir, path)
			route := "/" + strings.TrimSuffix(rel, ".tsx")
			if route == "/index" {
				route = "/"
			}
			layout := findNearestLayout(filepath.Dir(path), pagesDir)
			layoutRel := ""
			if layout != "" {
				layoutRel, _ = filepath.Rel(pagesDir, layout)
				layoutRel = filepath.ToSlash(layoutRel)
				layoutRel = "pages/" + layoutRel
			}
			pageRel, _ := filepath.Rel("../user-app", path)
			pageRel = filepath.ToSlash(pageRel)
			pageRel = "pages/" + strings.TrimPrefix(pageRel, "pages/")
			routes[route] = RouteInfo{
				PagePath:   pageRel,
				LayoutPath: layoutRel,
			}
		}
		return nil
	})
	return routes
}

func findNearestLayout(dir, pagesDir string) string {
	for {
		layoutPath := filepath.Join(dir, "layout.tsx")
		if fileExists(layoutPath) {
			return layoutPath
		}
		if dir == pagesDir || dir == "." || dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

func handlePageRenderWithLayout(w http.ResponseWriter, _ *http.Request, info RouteInfo) {
	props := map[string]interface{}{
		"frameworkName": "Go/React",
	}
	payload := map[string]interface{}{
		"componentPath": info.PagePath,
		"layoutPath":    info.LayoutPath,
		"props":         props,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to create render payload", http.StatusInternalServerError)
		return
	}
	resp, err := http.Post(nodeRendererURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to call node renderer: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read renderer response", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Node renderer returned an error: %s\n", string(body))
		http.Error(w, "Node renderer failed to render component", http.StatusInternalServerError)
		return
	}
	var renderResp struct {
		HTML     string                 `json:"html"`
		Error    string                 `json:"error,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}
	if err := json.Unmarshal(body, &renderResp); err != nil {
		http.Error(w, "Failed to decode renderer response", http.StatusInternalServerError)
		return
	}
	if renderResp.Error != "" {
		http.Error(w, fmt.Sprintf("Renderer error: %s", renderResp.Error), http.StatusInternalServerError)
		return
	}
	html := buildHTMLDocument(renderResp.HTML, props, info.PagePath, info.LayoutPath, renderResp.Metadata)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func buildHTMLDocument(content string, props map[string]interface{}, pagePath string, layoutPath string, metadata map[string]interface{}) string {
	propsJSON, _ := json.Marshal(props)
	pagePathJSON, _ := json.Marshal(pagePath)
	layoutPathJSON, _ := json.Marshal(layoutPath)

	title := "Go/React Framework"
	description := ""
	favicon := ""
	if metadata != nil {
		if t, ok := metadata["title"].(string); ok {
			title = t
		}
		if d, ok := metadata["description"].(string); ok {
			description = d
		}
		if f, ok := metadata["favicon"].(string); ok {
			favicon = f
		}
	}

	metaTags := ""
	if description != "" {
		metaTags += fmt.Sprintf("<meta name=\"description\" content=\"%s\">\n", description)
	}
	if favicon != "" {
		metaTags += fmt.Sprintf("<link rel=\"icon\" href=\"%s\">\n", favicon)
	}

	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + title + `</title>
    ` + metaTags + `
    <style>
        body { margin: 0; background-color: #282c34; color: white; }
    </style>
</head>
<body>
    <div id="root">` + content + `</div>
    <script>window.__SSR_PAGE__ = ` + string(pagePathJSON) + `; window.__SSR_LAYOUT__ = ` + string(layoutPathJSON) + `; window.__SSR_PROPS__ = ` + string(propsJSON) + `</script>
    <script src="/client.js" defer></script>
</body>
</html>
`
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Stub handler for Go server functions
func handleGoFunction(w http.ResponseWriter, r *http.Request) {
	functionName := chi.URLParam(r, "functionName")
	baseDir := "../user-app"
	err := serverfuncs.LoadAndCallGoPlugin(baseDir, functionName, w, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
	}
}

// Stub handler for TypeScript server functions
func handleTSFunction(w http.ResponseWriter, r *http.Request) {
	functionName := chi.URLParam(r, "functionName")

	// Parse request body
	var body interface{}
	if r.Body != nil {
		defer r.Body.Close()
		bodyBytes, _ := io.ReadAll(r.Body)
		if len(bodyBytes) > 0 {
			_ = json.Unmarshal(bodyBytes, &body)
		}
	}

	// Parse query params
	query := map[string][]string(r.URL.Query())

	// Collect headers
	headers := map[string][]string(r.Header)

	// Prepare input for the Node.js runner
	input := map[string]interface{}{
		"functionName": functionName,
		"body":         body,
		"query":        query,
		"headers":      headers,
	}
	inputBytes, _ := json.Marshal(input)

	cmd := exec.Command("node", "../node-renderer/ts-function-runner.js")
	cmd.Stdin = bytes.NewReader(inputBytes)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	output := outBuf.Bytes()
	if len(output) == 0 {
		output = errBuf.Bytes()
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(output)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(output)
}

func handleFunctionList(w http.ResponseWriter, r *http.Request) {
	baseDir := "../user-app"
	funcs, err := serverfuncs.DiscoverServerFunctions(baseDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(funcs)
}
