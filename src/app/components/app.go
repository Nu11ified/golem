//go:build js && wasm

package components

import (
	"github.com/Nu11ified/golem/css"
	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/router"
	"github.com/Nu11ified/golem/src/app/models"
	"github.com/Nu11ified/golem/src/app/store"
	"github.com/Nu11ified/golem/src/app/styles"
	"github.com/Nu11ified/golem/state"
)

// AppStore is the global Redux-style store
var AppStore *state.Store

// InitApp initializes the application: store, CSS, router
func InitApp() {
	// Inject global CSS
	css.InjectStyles(styles.GlobalCSS())

	// Create store with reducers and middleware
	AppStore = state.NewStore()
	AppStore.AddReducer("workspace", store.WorkspaceReducer, models.NewWorkspace())
	AppStore.AddReducer("editor", store.EditorReducer, models.NewEditorState())
	AppStore.AddMiddleware(store.LocalStorageMiddleware)
	AppStore.AddMiddleware(store.ServerSyncMiddleware)

	// Load from localStorage or create initial state
	p := state.NewPersistence()
	var ws models.Workspace
	err := p.LoadState("notion_workspace", &ws)
	if err != nil {
		// No saved state found; create default workspace with one page
		workspace := models.NewWorkspace()
		page := models.NewPage("Getting Started", "")
		workspace.Pages[page.ID] = page
		workspace.RootPageIDs = append(workspace.RootPageIDs, page.ID)
		AppStore.Dispatch(state.Action{
			Type:    "INIT_WORKSPACE",
			Payload: workspace,
		})
	}

	// Set up router
	r := router.NewRouter()
	r.AddSimpleRoute("/:pageId", func(params map[string]string) *dom.Element {
		return EditorPage(params["pageId"])
	})
	r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
		return RedirectToFirstPage()
	})
	r.Start()
}

// RedirectToFirstPage navigates to the first available page
func RedirectToFirstPage() *dom.Element {
	return dom.Div(dom.Text("Loading..."))
}

// EditorPage is a placeholder -- will be implemented in Task 14
func EditorPage(pageID string) *dom.Element {
	return dom.Div(dom.Class("editor-page"), dom.Text("Editor: "+pageID))
}
