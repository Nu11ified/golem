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

	// Set up router using the default router so that package-level
	// router.Navigate() calls (used by RedirectToFirstPage, Sidebar,
	// etc.) go through the same instance.
	router.DefaultRouter.SetContainer("#app")
	router.AddRoute("/:pageId", func(params map[string]string) *dom.Element {
		return EditorPage(params["pageId"])
	})
	router.AddRoute("/", func(params map[string]string) *dom.Element {
		return RedirectToFirstPage()
	})
	router.Start()
}

// RedirectToFirstPage navigates to the first available page
func RedirectToFirstPage() *dom.Element {
	ws, ok := AppStore.GetState("workspace").(*models.Workspace)
	if ok && len(ws.RootPageIDs) > 0 {
		router.Navigate("/" + ws.RootPageIDs[0])
		return dom.Div(dom.Text("Redirecting..."))
	}
	// No pages yet — create one
	AppStore.Dispatch(state.Action{
		Type:    store.CreatePage,
		Payload: map[string]interface{}{"title": "Getting Started", "parentId": ""},
	})
	ws, _ = AppStore.GetState("workspace").(*models.Workspace)
	if len(ws.RootPageIDs) > 0 {
		router.Navigate("/" + ws.RootPageIDs[0])
	}
	return dom.Div(dom.Text("Redirecting..."))
}

// EditorPage renders the full editor for a given page
func EditorPage(pageID string) *dom.Element {
	ws, _ := AppStore.GetState("workspace").(*models.Workspace)
	if ws == nil {
		return dom.Div(dom.Text("Loading..."))
	}
	page, ok := ws.Pages[pageID]
	if !ok {
		return dom.Div(dom.Text("Page not found"))
	}

	// Load or create editor state for this page
	es, _ := AppStore.GetState("editor").(*models.EditorState)
	if es == nil || es.ActivePageID != pageID {
		es = models.NewEditorState()
		es.ActivePageID = pageID
	}

	sidebar := Sidebar(ws, pageID)
	editor := RenderEditorPage(page, es)
	return AppLayout(sidebar, editor)
}
