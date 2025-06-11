package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const nodeRendererURL = "http://localhost:3001/render"

type RenderPayload struct {
	ComponentPath string                 `json:"componentPath"`
	Props         map[string]interface{} `json:"props"`
}

type RenderResponse struct {
	HTML  string `json:"html"`
	Error string `json:"error,omitempty"`
}

// RouteInfo holds the mapping from route to page and layout
type RouteInfo struct {
	PagePath   string
	LayoutPath string
}

var routeMap map[string]RouteInfo

func main() {
	routeMap = buildRouteMap("../user-app/pages")
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Serve static files (client.js, client.js.map)
	r.Handle("/client.js", http.FileServer(http.Dir("../node-renderer/dist")))
	r.Handle("/client.js.map", http.FileServer(http.Dir("../node-renderer/dist")))

	// Handle favicon.ico requests with 404
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	// Handle .well-known and other static paths with 404
	r.Handle("/.well-known/*", http.HandlerFunc(http.NotFound))

	// Serve static assets from user-app/public if they exist
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		publicPath := "../user-app/public" + r.URL.Path
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

	log.Println("Go orchestrator listening on http://localhost:8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// buildRouteMap recursively scans the pages directory and builds a map of routes to page/layout
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

// findNearestLayout walks up the directory tree to find the nearest layout.tsx
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

// handlePageRenderWithLayout passes both page and layout to the Node renderer
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
		log.Printf("Node renderer returned an error: %s", string(body))
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
    <script src="/client.js"></script>
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
