package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
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

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/*", handlePageRender)

	log.Println("Go orchestrator listening on http://localhost:8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handlePageRender(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path

	var componentPath string
	if requestPath == "/" {
		componentPath = "pages/index.tsx"
	} else {
		// Basic routing: /about -> pages/about.tsx
		// For this PoC, we only support .tsx extensions implicitly.
		componentPath = path.Join("pages", strings.Trim(requestPath, "/")+".tsx")
	}

	// For this PoC, we'll pass a static prop.
	// A real implementation would fetch data here based on the request.
	props := map[string]interface{}{
		"frameworkName": "Go/React",
	}

	payload := RenderPayload{
		ComponentPath: componentPath,
		Props:         props,
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

	var renderResp RenderResponse
	if err := json.Unmarshal(body, &renderResp); err != nil {
		http.Error(w, "Failed to decode renderer response", http.StatusInternalServerError)
		return
	}

	if renderResp.Error != "" {
		http.Error(w, fmt.Sprintf("Renderer error: %s", renderResp.Error), http.StatusInternalServerError)
		return
	}

	html := buildHTMLDocument(renderResp.HTML, props)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func buildHTMLDocument(content string, props map[string]interface{}) string {
	propsJSON, _ := json.Marshal(props)
	return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go/React Framework</title>
    <style>
        body { margin: 0; background-color: #282c34; color: white; }
    </style>
</head>
<body>
    <div id="root">` + content + `</div>
    <script>window.__SSR_PROPS__ = ` + string(propsJSON) + `</script>
    <script src="http://localhost:3001/client.js"></script>
</body>
</html>
`
}
