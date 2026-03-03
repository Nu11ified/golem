//go:build !js || !wasm

package components

import (
	"encoding/json"
	"fmt"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/src/app/models"
	"github.com/Nu11ified/golem/src/app/styles"
)

// RenderFullPage renders the complete page layout for SSR
func RenderFullPage(ws *models.Workspace, page *models.Page, es *models.EditorState) *dom.Element {
	sidebar := Sidebar(ws, page.ID)
	editor := RenderEditorPage(page, es)
	return AppLayout(sidebar, editor)
}

// GenerateSSRDocument generates a complete HTML document for SSR
func GenerateSSRDocument(ws *models.Workspace, page *models.Page, es *models.EditorState) string {
	// Render the app content
	content := RenderFullPage(ws, page, es)
	bodyHTML := dom.RenderToHTML(content)

	// Serialize initial state
	initialState := map[string]interface{}{
		"workspace": ws,
		"editor":    es,
	}
	stateJSON, _ := json.Marshal(initialState)

	// Build the full HTML document
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>%s</style>
</head>
<body>
    <div id="app">%s</div>
    <script id="__GOLEM_STATE__" type="application/json">%s</script>
    <script src="/wasm_exec.js"></script>
    <script>
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("/app.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        });
    </script>
</body>
</html>`, page.Title, styles.GlobalCSS(), bodyHTML, string(stateJSON))
}
