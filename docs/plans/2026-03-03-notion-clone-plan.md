# Notion Clone Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Notion-style workspace app with page tree sidebar, block editor, localStorage + server sync, and SSR with hydration — all using the Golem Go/WASM framework.

**Architecture:** Document-centric Redux-style Store with "workspace" and "editor" slices. localStorage middleware persists immediately; server sync middleware debounces writes. SSR renders the initial page with embedded state; WASM hydrates. Dual-file pattern (WASM + stub) for all shared types.

**Tech Stack:** Golem framework (Go/WASM), `state.Store` + reducers, `dom.*` element API, `css.InjectStyles`, file-based routing with `router.*`, server functions via `functions.Register`, SSR via `internal/ssr`, JSON file storage on server.

---

### Task 1: Data Model Types

**Files:**
- Create: `src/app/models/types.go` (WASM build)
- Create: `src/app/models/types_stub.go` (non-WASM build, same types)
- Test: `test/unit/models/types_test.go`

**Context:** All types are shared between WASM client and server-side SSR. Both files define identical structs but with different build tags. The types are pure data — no syscall/js dependencies.

**Step 1: Write the test file**

```go
//go:build !js || !wasm

package models_test

import (
    "encoding/json"
    "testing"

    "github.com/Nu11ified/golem/src/app/models"
)

func TestPageJSON(t *testing.T) {
    p := &models.Page{
        ID:       "page-1",
        Title:    "Test Page",
        ParentID: "",
        ChildIDs: []string{"page-2"},
        Icon:     "📄",
    }
    data, err := json.Marshal(p)
    if err != nil {
        t.Fatal(err)
    }
    var p2 models.Page
    if err := json.Unmarshal(data, &p2); err != nil {
        t.Fatal(err)
    }
    if p2.ID != "page-1" || p2.Title != "Test Page" || p2.Icon != "📄" {
        t.Errorf("round-trip mismatch: %+v", p2)
    }
    if len(p2.ChildIDs) != 1 || p2.ChildIDs[0] != "page-2" {
        t.Errorf("childIDs mismatch: %v", p2.ChildIDs)
    }
}

func TestBlockJSON(t *testing.T) {
    b := &models.Block{
        ID:      "block-1",
        Type:    "text",
        Content: "Hello world",
        PageID:  "page-1",
    }
    data, err := json.Marshal(b)
    if err != nil {
        t.Fatal(err)
    }
    var b2 models.Block
    if err := json.Unmarshal(data, &b2); err != nil {
        t.Fatal(err)
    }
    if b2.Type != "text" || b2.Content != "Hello world" {
        t.Errorf("round-trip mismatch: %+v", b2)
    }
}

func TestNewWorkspace(t *testing.T) {
    ws := models.NewWorkspace()
    if ws.Pages == nil {
        t.Fatal("Pages map should be initialized")
    }
    if ws.RootPageIDs == nil {
        t.Fatal("RootPageIDs should be initialized")
    }
}

func TestNewEditorState(t *testing.T) {
    es := models.NewEditorState()
    if es.Blocks == nil {
        t.Fatal("Blocks map should be initialized")
    }
    if es.BlockOrder == nil {
        t.Fatal("BlockOrder should be initialized")
    }
}

func TestNewPage(t *testing.T) {
    p := models.NewPage("Test", "")
    if p.ID == "" {
        t.Fatal("ID should be generated")
    }
    if p.Title != "Test" {
        t.Errorf("expected title 'Test', got %q", p.Title)
    }
    if p.CreatedAt == 0 {
        t.Error("CreatedAt should be set")
    }
}

func TestNewBlock(t *testing.T) {
    b := models.NewBlock("text", "page-1")
    if b.ID == "" {
        t.Fatal("ID should be generated")
    }
    if b.Type != "text" || b.PageID != "page-1" {
        t.Errorf("unexpected block: %+v", b)
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./test/unit/models/ -v`
Expected: FAIL — package not found

**Step 3: Write the non-WASM stub (used by tests and SSR)**

Create `src/app/models/types_stub.go`:

```go
//go:build !js || !wasm

package models

import (
    "fmt"
    "time"
)

type Page struct {
    ID        string   `json:"id"`
    Title     string   `json:"title"`
    ParentID  string   `json:"parentId"`
    ChildIDs  []string `json:"childIds"`
    Icon      string   `json:"icon"`
    CreatedAt int64    `json:"createdAt"`
    UpdatedAt int64    `json:"updatedAt"`
}

type Block struct {
    ID       string                 `json:"id"`
    Type     string                 `json:"type"`
    Content  string                 `json:"content"`
    Children []string               `json:"children"`
    Props    map[string]interface{} `json:"props"`
    PageID   string                 `json:"pageId"`
}

type Workspace struct {
    Pages       map[string]*Page `json:"pages"`
    RootPageIDs []string         `json:"rootPageIds"`
}

type EditorState struct {
    ActivePageID string           `json:"activePageId"`
    Blocks       map[string]*Block `json:"blocks"`
    BlockOrder   []string          `json:"blockOrder"`
    FocusBlockID string            `json:"focusBlockId"`
}

func NewWorkspace() *Workspace {
    return &Workspace{
        Pages:       make(map[string]*Page),
        RootPageIDs: make([]string, 0),
    }
}

func NewEditorState() *EditorState {
    return &EditorState{
        Blocks:     make(map[string]*Block),
        BlockOrder: make([]string, 0),
    }
}

var idCounter int

func generateID() string {
    idCounter++
    return fmt.Sprintf("%d-%d", time.Now().UnixMilli(), idCounter)
}

func NewPage(title, parentID string) *Page {
    now := time.Now().UnixMilli()
    return &Page{
        ID:        generateID(),
        Title:     title,
        ParentID:  parentID,
        ChildIDs:  make([]string, 0),
        Icon:      "📄",
        CreatedAt: now,
        UpdatedAt: now,
    }
}

func NewBlock(blockType, pageID string) *Block {
    return &Block{
        ID:       generateID(),
        Type:     blockType,
        PageID:   pageID,
        Children: make([]string, 0),
        Props:    make(map[string]interface{}),
    }
}
```

Then create `src/app/models/types.go` — identical content but with `//go:build js && wasm` build tag.

**Step 4: Run tests**

Run: `go test ./test/unit/models/ -v`
Expected: PASS (6 tests)

**Step 5: Commit**

```bash
git add src/app/models/ test/unit/models/
git commit -m "feat(notion): add data model types for Page, Block, Workspace, EditorState"
```

---

### Task 2: Workspace Reducer

**Files:**
- Create: `src/app/store/workspace.go` (non-WASM, same code works for both)
- Create: `src/app/store/workspace_stub.go` (stub with build tag)
- Test: `test/unit/store/workspace_test.go`

**Context:** The workspace reducer manages the page tree. It handles page CRUD, reparenting, and reordering. Reducer signature: `func(state interface{}, action state.Action) interface{}`. State is `*models.Workspace`. Actions use string constants.

**Step 1: Write the test file**

```go
//go:build !js || !wasm

package store_test

import (
    "testing"

    "github.com/Nu11ified/golem/src/app/models"
    "github.com/Nu11ified/golem/src/app/store"
    golemstate "github.com/Nu11ified/golem/state"
)

func TestCreatePage(t *testing.T) {
    ws := models.NewWorkspace()
    action := golemstate.Action{
        Type:    store.CreatePage,
        Payload: map[string]interface{}{"title": "New Page", "parentId": ""},
    }
    result := store.WorkspaceReducer(ws, action)
    ws2 := result.(*models.Workspace)
    if len(ws2.RootPageIDs) != 1 {
        t.Fatalf("expected 1 root page, got %d", len(ws2.RootPageIDs))
    }
    page := ws2.Pages[ws2.RootPageIDs[0]]
    if page.Title != "New Page" {
        t.Errorf("expected title 'New Page', got %q", page.Title)
    }
}

func TestDeletePage(t *testing.T) {
    ws := models.NewWorkspace()
    // Create a page first
    ws = store.WorkspaceReducer(ws, golemstate.Action{
        Type:    store.CreatePage,
        Payload: map[string]interface{}{"title": "To Delete", "parentId": ""},
    }).(*models.Workspace)
    pageID := ws.RootPageIDs[0]

    // Delete it
    ws = store.WorkspaceReducer(ws, golemstate.Action{
        Type:    store.DeletePage,
        Payload: pageID,
    }).(*models.Workspace)

    if len(ws.RootPageIDs) != 0 {
        t.Errorf("expected 0 root pages, got %d", len(ws.RootPageIDs))
    }
    if _, exists := ws.Pages[pageID]; exists {
        t.Error("page should be deleted from map")
    }
}

func TestRenamePage(t *testing.T) {
    ws := models.NewWorkspace()
    ws = store.WorkspaceReducer(ws, golemstate.Action{
        Type:    store.CreatePage,
        Payload: map[string]interface{}{"title": "Old", "parentId": ""},
    }).(*models.Workspace)
    pageID := ws.RootPageIDs[0]

    ws = store.WorkspaceReducer(ws, golemstate.Action{
        Type:    store.RenamePage,
        Payload: map[string]interface{}{"id": pageID, "title": "New"},
    }).(*models.Workspace)

    if ws.Pages[pageID].Title != "New" {
        t.Errorf("expected 'New', got %q", ws.Pages[pageID].Title)
    }
}

func TestCreateChildPage(t *testing.T) {
    ws := models.NewWorkspace()
    ws = store.WorkspaceReducer(ws, golemstate.Action{
        Type:    store.CreatePage,
        Payload: map[string]interface{}{"title": "Parent", "parentId": ""},
    }).(*models.Workspace)
    parentID := ws.RootPageIDs[0]

    ws = store.WorkspaceReducer(ws, golemstate.Action{
        Type:    store.CreatePage,
        Payload: map[string]interface{}{"title": "Child", "parentId": parentID},
    }).(*models.Workspace)

    if len(ws.RootPageIDs) != 1 {
        t.Fatalf("child should not be in root pages")
    }
    parent := ws.Pages[parentID]
    if len(parent.ChildIDs) != 1 {
        t.Fatalf("expected 1 child, got %d", len(parent.ChildIDs))
    }
    child := ws.Pages[parent.ChildIDs[0]]
    if child.Title != "Child" || child.ParentID != parentID {
        t.Errorf("unexpected child: %+v", child)
    }
}

func TestSetPageIcon(t *testing.T) {
    ws := models.NewWorkspace()
    ws = store.WorkspaceReducer(ws, golemstate.Action{
        Type:    store.CreatePage,
        Payload: map[string]interface{}{"title": "Test", "parentId": ""},
    }).(*models.Workspace)
    pageID := ws.RootPageIDs[0]

    ws = store.WorkspaceReducer(ws, golemstate.Action{
        Type:    store.SetPageIcon,
        Payload: map[string]interface{}{"id": pageID, "icon": "🚀"},
    }).(*models.Workspace)

    if ws.Pages[pageID].Icon != "🚀" {
        t.Errorf("expected icon '🚀', got %q", ws.Pages[pageID].Icon)
    }
}
```

**Step 2: Run tests — expected FAIL**

**Step 3: Implement workspace reducer**

Create `src/app/store/workspace_stub.go` with `//go:build !js || !wasm`:
- Action type constants: `CreatePage`, `DeletePage`, `RenamePage`, `MovePage`, `SetPageIcon`, `ReorderPages`
- `WorkspaceReducer` function matching the `state.Reducer` signature
- Switch on `action.Type`, cast `state` to `*models.Workspace`, perform mutation, return

Create `src/app/store/workspace.go` with `//go:build js && wasm` — identical content.

**Step 4: Run tests — expected PASS (5 tests)**

**Step 5: Commit**

```bash
git commit -m "feat(notion): add workspace reducer with page CRUD"
```

---

### Task 3: Editor Reducer

**Files:**
- Create: `src/app/store/editor.go` + `src/app/store/editor_stub.go`
- Test: `test/unit/store/editor_test.go`

**Context:** The editor reducer manages blocks for the active page. Same dual-file pattern as workspace.

**Step 1: Write tests**

Test: `LoadPage` (sets activePageID, loads blocks), `AddBlock` (inserts at position), `DeleteBlock` (removes and updates order), `UpdateBlock` (changes content), `ChangeBlockType` (converts text to h1 etc.), `MoveBlock` (reorders), `ToggleBlock` (expands/collapses).

**Step 2: Implement**

Action constants: `LoadPage`, `AddBlock`, `DeleteBlock`, `UpdateBlock`, `MoveBlock`, `ToggleBlock`, `ChangeBlockType`.

`EditorReducer(state interface{}, action state.Action) interface{}`:
- `LoadPage`: payload is `map[string]interface{}{"pageId": "...", "blocks": []*Block, "blockOrder": []string}` — replaces entire editor state
- `AddBlock`: payload has `blockType`, `afterBlockID`, `pageId` — creates a new block, inserts into BlockOrder after the specified block
- `DeleteBlock`: payload is block ID — removes from Blocks map and BlockOrder
- `UpdateBlock`: payload has `id`, `content` — updates block content
- `ChangeBlockType`: payload has `id`, `newType` — changes block Type field
- `MoveBlock`: payload has `id`, `newIndex` — repositions in BlockOrder
- `ToggleBlock`: payload is block ID — toggles a boolean prop `collapsed`

**Step 3: Run tests — expected PASS**

**Step 4: Commit**

```bash
git commit -m "feat(notion): add editor reducer with block CRUD"
```

---

### Task 4: Server-Side Persistence & Functions

**Files:**
- Create: `src/server/storage.go` (JSON file storage)
- Create: `src/server/notion.go` (server function registration)
- Test: `test/unit/server/storage_test.go`

**Context:** No build tag needed — server code is host-only. Uses `functions.Register` in `init()`.

**Step 1: Write tests for file storage**

```go
package server_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/Nu11ified/golem/src/server"
)

func TestSaveAndLoadWorkspace(t *testing.T) {
    dir := t.TempDir()
    s := server.NewStorage(dir)

    ws := map[string]interface{}{
        "pages":       map[string]interface{}{},
        "rootPageIds": []interface{}{},
    }
    if err := s.SaveWorkspace(ws); err != nil {
        t.Fatal(err)
    }

    loaded, err := s.LoadWorkspace()
    if err != nil {
        t.Fatal(err)
    }
    if loaded == nil {
        t.Fatal("expected non-nil workspace")
    }
}

func TestSaveAndLoadBlocks(t *testing.T) {
    dir := t.TempDir()
    s := server.NewStorage(dir)

    blocks := []interface{}{
        map[string]interface{}{"id": "b1", "type": "text", "content": "Hello"},
    }
    if err := s.SavePageBlocks("page-1", blocks); err != nil {
        t.Fatal(err)
    }

    loaded, err := s.LoadPageBlocks("page-1")
    if err != nil {
        t.Fatal(err)
    }
    if len(loaded) != 1 {
        t.Fatalf("expected 1 block, got %d", len(loaded))
    }
}

func TestDeletePageData(t *testing.T) {
    dir := t.TempDir()
    s := server.NewStorage(dir)

    s.SavePageBlocks("page-1", []interface{}{})
    s.DeletePageData("page-1")

    blocksFile := filepath.Join(dir, "blocks", "page-1.json")
    if _, err := os.Stat(blocksFile); !os.IsNotExist(err) {
        t.Error("blocks file should be deleted")
    }
}
```

**Step 2: Implement Storage**

`src/server/storage.go`:
- `type Storage struct { dataDir string }`
- `NewStorage(dataDir string) *Storage`
- `SaveWorkspace(data interface{}) error` — writes JSON to `<dataDir>/workspace.json`
- `LoadWorkspace() (interface{}, error)` — reads and returns parsed JSON
- `SavePageBlocks(pageID string, blocks interface{}) error` — writes to `<dataDir>/blocks/<pageID>.json`
- `LoadPageBlocks(pageID string) ([]interface{}, error)`
- `DeletePageData(pageID string) error` — removes the blocks file

**Step 3: Register server functions**

`src/server/notion.go`:
```go
package server

import "github.com/Nu11ified/golem/functions"

var store *Storage

func init() {
    store = NewStorage("data")
    functions.Register("notion", "GetWorkspace", store.GetWorkspace)
    functions.Register("notion", "SaveWorkspace", store.SaveWorkspaceRPC)
    functions.Register("notion", "GetPageBlocks", store.GetPageBlocks)
    functions.Register("notion", "SavePageBlocks", store.SavePageBlocksRPC)
    functions.Register("notion", "DeletePageData", store.DeletePageDataRPC)
}

// RPC wrappers with simpler signatures for the function registry
func (s *Storage) GetWorkspace() (interface{}, error) { return s.LoadWorkspace() }
func (s *Storage) SaveWorkspaceRPC(data interface{}) error { return s.SaveWorkspace(data) }
func (s *Storage) GetPageBlocks(pageID string) ([]interface{}, error) { return s.LoadPageBlocks(pageID) }
func (s *Storage) SavePageBlocksRPC(pageID string, blocks interface{}) error { return s.SavePageBlocks(pageID, blocks) }
func (s *Storage) DeletePageDataRPC(pageID string) error { return s.DeletePageData(pageID) }
```

**Step 4: Run tests — expected PASS**

**Step 5: Commit**

```bash
git commit -m "feat(notion): add server-side JSON file storage and RPC functions"
```

---

### Task 5: Store Middleware (localStorage + Server Sync)

**Files:**
- Create: `src/app/store/middleware.go` (WASM — real localStorage + server calls)
- Create: `src/app/store/middleware_stub.go` (non-WASM — no-ops for testing)
- Test: `test/unit/store/middleware_test.go`

**Context:** Two middleware functions:
1. `LocalStorageMiddleware` — after each dispatch, serializes workspace/editor state to localStorage
2. `ServerSyncMiddleware` — debounces 500ms, then calls server functions via `grpc.Call`

The stub version records calls for testing but doesn't use `syscall/js` or `grpc`.

**Step 1: Write tests using the stub**

Test that middleware calls `next(action)` (doesn't break the chain), and that the stub records persist/sync calls.

**Step 2: Implement stubs and WASM versions**

The WASM `LocalStorageMiddleware`:
```go
func LocalStorageMiddleware(store *golemstate.Store, action golemstate.Action, next func(golemstate.Action)) {
    next(action)
    // After reducers have run, persist to localStorage
    ws := store.GetState("workspace")
    ed := store.GetState("editor")
    p := golemstate.NewPersistence()
    p.SaveState("notion_workspace", ws)
    p.SaveState("notion_editor", ed)
}
```

The WASM `ServerSyncMiddleware` uses a debounced goroutine with `time.AfterFunc(500ms)` that calls `grpc.Call(ctx, "notion", "SaveWorkspace", ws)`.

**Step 3: Commit**

```bash
git commit -m "feat(notion): add localStorage and server sync middleware"
```

---

### Task 6: CSS Design System

**Files:**
- Create: `src/app/styles/styles.go` (WASM — injects CSS)
- Create: `src/app/styles/styles_stub.go` (non-WASM — returns CSS string for SSR)
- Test: `test/unit/styles/styles_test.go`

**Context:** Global CSS with Notion-style aesthetic. CSS variables for theming. Responsive breakpoints. Tests verify CSS string contains expected rules.

**Step 1: Write tests**

```go
func TestCSSContainsSidebarStyles(t *testing.T) {
    css := styles.GlobalCSS()
    if !strings.Contains(css, ".sidebar") {
        t.Error("missing .sidebar class")
    }
    if !strings.Contains(css, "--bg-sidebar") {
        t.Error("missing --bg-sidebar variable")
    }
}

func TestCSSContainsEditorStyles(t *testing.T) {
    css := styles.GlobalCSS()
    if !strings.Contains(css, ".editor-area") {
        t.Error("missing .editor-area class")
    }
    if !strings.Contains(css, ".block") {
        t.Error("missing .block class")
    }
}

func TestCSSContainsMobileStyles(t *testing.T) {
    css := styles.GlobalCSS()
    if !strings.Contains(css, "@media") {
        t.Error("missing media queries")
    }
    if !strings.Contains(css, "768px") {
        t.Error("missing 768px breakpoint")
    }
}
```

**Step 2: Implement**

`styles.GlobalCSS() string` returns the full CSS string. The WASM version also calls `css.InjectStyles(GlobalCSS())` at init. CSS includes:
- `:root` variables (colors, fonts, spacing)
- `.app-layout` (flex, full-height)
- `.sidebar` (240px, bg-sidebar, overflow-y, transition for mobile slide)
- `.sidebar-header`, `.page-tree`, `.page-tree-item` (indentation, hover, active)
- `.editor-area` (flex-grow, centered, max-width 720px)
- `.page-header` (icon + title, large font)
- `.block` (contenteditable styling, focus border, hover highlight)
- `.block-text`, `.block-h1`, `.block-h2`, `.block-h3`, `.block-bullet`, `.block-numbered`, `.block-toggle`, `.block-code`, `.block-divider`
- `.slash-menu` (absolute positioned dropdown)
- `.sidebar-overlay`, `.hamburger-btn` (mobile drawer)
- `@media (max-width: 768px)` (sidebar hidden by default, overlay, hamburger visible)
- Touch target sizing (min 44px)

**Step 3: Commit**

```bash
git commit -m "feat(notion): add CSS design system with Notion aesthetic and mobile responsive"
```

---

### Task 7: App Layout & Routing

**Files:**
- Create: `src/app/components/layout.go` + `layout_stub.go`
- Create: `src/app/components/app.go` + `app_stub.go`
- Modify: `src/app/main.go` (wire up router, store, CSS)
- Test: `test/unit/components/layout_test.go`

**Context:** The layout wraps the sidebar (parallel slot) and editor area. The router has a dynamic `/:pageId` route. The app entry point initializes the store, loads from localStorage, registers routes, and either hydrates (if SSR) or renders.

**Step 1: Write tests**

Test that AppLayout returns an element with class `app-layout`, and that it contains both `.sidebar` and `.editor-area` children.

**Step 2: Implement**

`AppLayout(sidebar, content *dom.Element) *dom.Element`:
```go
func AppLayout(sidebar, content *dom.Element) *dom.Element {
    return dom.Div(dom.Class("app-layout"),
        sidebar,
        dom.Div(dom.Class("editor-area"), content),
    )
}
```

Route setup (in the WASM main.go):
```go
r := router.NewRouter()
r.SetMode(router.HistoryMode)
r.SetContainer("#app")

// Register routes
r.AddRoute(&router.Route{
    Path: "/:pageId",
    Component: func(params map[string]string) *dom.Element {
        return EditorPage(params["pageId"])
    },
    Layout: func(child *dom.Element) *dom.Element {
        return AppLayout(Sidebar(), child)
    },
})

r.AddSimpleRoute("/", func(params map[string]string) *dom.Element {
    // Redirect to first page or create one
    return RedirectToFirstPage()
})
```

**Step 3: Commit**

```bash
git commit -m "feat(notion): add app layout, routing, and entry point"
```

---

### Task 8: Sidebar Components

**Files:**
- Create: `src/app/components/sidebar.go` + `sidebar_stub.go`
- Create: `src/app/components/page_tree.go` + `page_tree_stub.go`
- Test: `test/unit/components/sidebar_test.go`

**Context:** Sidebar has a workspace header (title + "new page" button), search bar, and recursive page tree. Tree items are indented by depth, show icon + title, click to navigate.

**Step 1: Write tests**

Test that `Sidebar()` returns element with class `sidebar`. Test that `PageTree(workspace, activePageID)` renders correct number of items. Test that nested pages are indented (have depth class).

**Step 2: Implement**

`Sidebar() *dom.Element` — reads workspace state from store, builds the sidebar. WASM version subscribes to store changes for re-rendering.

`PageTree(ws *models.Workspace, activeID string) *dom.Element` — iterates `RootPageIDs`, recursively renders `PageTreeItem` for each page and its children.

`PageTreeItem(page *models.Page, depth int, activeID string, ws *models.Workspace) *dom.Element`:
- `dom.Div(dom.Class("page-tree-item"), style for indent)`
- Contains icon span + title span
- WASM: `dom.OnClick` navigates to `/<pageId>`
- Active item gets `page-tree-item-active` class
- Recursively renders children at `depth+1`

"New page" button dispatches `CreatePage` action, then navigates to the new page.

**Step 3: Commit**

```bash
git commit -m "feat(notion): add sidebar with page tree and navigation"
```

---

### Task 9: Page Header Component

**Files:**
- Create: `src/app/components/page_header.go` + `page_header_stub.go`
- Test: `test/unit/components/page_header_test.go`

**Context:** Shows the page icon (clickable to change) and an editable title. Title changes dispatch `RenamePage`. Icon changes dispatch `SetPageIcon`.

**Step 1: Write tests**

Test that `PageHeader(page)` renders with the page title. Test that the icon is present.

**Step 2: Implement**

`PageHeader(page *models.Page) *dom.Element`:
- Icon span (emoji) — WASM: `dom.OnClick` toggles an emoji picker
- Title: `dom.Input` with page title as value, `dom.OnInput` dispatches `RenamePage`
- Style: large font, no border on input, Notion-style placeholder "Untitled"

Emoji picker: a simple grid of common emojis shown/hidden on click. Selecting one dispatches `SetPageIcon`.

**Step 3: Commit**

```bash
git commit -m "feat(notion): add page header with editable title and icon picker"
```

---

### Task 10: Block Renderer & Basic Block Types

**Files:**
- Create: `src/app/components/blocks/renderer.go` + `renderer_stub.go`
- Create: `src/app/components/blocks/text.go` + `text_stub.go`
- Create: `src/app/components/blocks/heading.go` + `heading_stub.go`
- Create: `src/app/components/blocks/divider.go` + `divider_stub.go`
- Test: `test/unit/components/blocks/renderer_test.go`

**Context:** `BlockRenderer` dispatches to the correct block component based on `block.Type`. Each block is a `contenteditable` div. WASM versions use `js.Global()` for `contenteditable` events since Golem only has `click`, `input`, `change`, `keydown` pre-wired.

**Step 1: Write tests**

Test that `RenderBlock(block)` returns correct element type for each block type. Test that text block contains the content. Test that heading blocks use correct tag (h1, h2, h3). Test that divider renders an `<hr>`.

**Step 2: Implement**

`RenderBlock(block *models.Block) *dom.Element` — switch on `block.Type`, delegate to specific render function.

`TextBlock(block *models.Block) *dom.Element`:
- SSR stub: `dom.P(dom.Class("block block-text"), dom.Text(block.Content))`
- WASM: creates a `contenteditable` div, attaches keydown/input handlers via `js.Global()`:
  - On input: dispatch `UpdateBlock` with new content
  - On Enter: dispatch `AddBlock` after this block
  - On Backspace when empty: dispatch `DeleteBlock`
  - On typing `# ` at start: dispatch `ChangeBlockType` to h1

`HeadingBlock(block *models.Block) *dom.Element`:
- Uses `dom.El("h1"/"h2"/"h3")` based on block.Type
- Same contenteditable pattern as text but with heading class

`DividerBlock(block *models.Block) *dom.Element`:
- `dom.El("hr", dom.Class("block block-divider"))`

**Step 3: Commit**

```bash
git commit -m "feat(notion): add block renderer with text, heading, and divider blocks"
```

---

### Task 11: List, Toggle, and Code Blocks

**Files:**
- Create: `src/app/components/blocks/list.go` + `list_stub.go`
- Create: `src/app/components/blocks/toggle.go` + `toggle_stub.go`
- Create: `src/app/components/blocks/code.go` + `code_stub.go`
- Test: `test/unit/components/blocks/list_test.go`

**Context:** Extends block renderer with remaining types.

**Step 1: Write tests**

Test bullet list renders with bullet marker. Test numbered list renders with number. Test toggle renders with expand/collapse indicator. Test code block renders with monospace font class.

**Step 2: Implement**

`BulletListBlock`: contenteditable div with a bullet `•` prefix span.

`NumberedListBlock`: contenteditable div with a number prefix. The number is computed from position in BlockOrder among consecutive numbered blocks.

`ToggleBlock`: a triangle indicator `▶`/`▼` + contenteditable content. Clicking the triangle dispatches `ToggleBlock`. When expanded, renders child blocks (recursively via `RenderBlock`).

`CodeBlock`: a `<pre><code>` wrapper with class `block-code`. Content is plain text, not contenteditable rich text. Uses a `<textarea>` internally for editing in WASM.

**Step 3: Commit**

```bash
git commit -m "feat(notion): add list, toggle, and code block types"
```

---

### Task 12: Block Editing Interactions

**Files:**
- Create: `src/app/components/blocks/interactions.go` (WASM only)
- Modify: block type files to use shared interaction handlers
- Test: `test/unit/components/blocks/interactions_test.go` (test the type-conversion logic)

**Context:** Block-level keyboard shortcuts. These are WASM-only (contenteditable events via `js.Global()`). The non-WASM stubs don't need these since SSR doesn't handle editing.

**Step 1: Write tests for type detection**

Test pure functions (no JS dependency):
- `DetectBlockType("# Hello")` → `("h1", "Hello")`
- `DetectBlockType("## Hello")` → `("h2", "Hello")`
- `DetectBlockType("### Hello")` → `("h3", "Hello")`
- `DetectBlockType("- Hello")` → `("bullet", "Hello")`
- `DetectBlockType("* Hello")` → `("bullet", "Hello")`
- `DetectBlockType("1. Hello")` → `("numbered", "Hello")`
- `DetectBlockType("---")` → `("divider", "")`
- `DetectBlockType("```")` → `("code", "")`
- `DetectBlockType("Hello")` → `("", "")` (no conversion)

**Step 2: Implement**

`DetectBlockType(text string) (newType, remainingContent string)` — pure string parsing, no JS.

`SetupBlockKeyHandler(block *models.Block, element js.Value, store *golemstate.Store)` — WASM-only function that:
1. Adds `keydown` event listener
2. On Enter: prevents default, gets cursor position, splits content, dispatches `AddBlock` with remaining text
3. On Backspace when content is empty: prevents default, dispatches `DeleteBlock`, focuses previous block
4. On any input: checks `DetectBlockType`, dispatches `ChangeBlockType` if matched, sets remaining content

`FocusBlock(blockID string)` — WASM-only, uses `document.querySelector` to find and focus the block element.

**Step 3: Commit**

```bash
git commit -m "feat(notion): add block editing interactions and type conversion"
```

---

### Task 13: Slash Command Menu

**Files:**
- Create: `src/app/components/slash_menu.go` + `slash_menu_stub.go`
- Test: `test/unit/components/slash_menu_test.go`

**Context:** Typing `/` in a block opens a menu with block type options. Selecting one changes the current block's type. The menu is positioned below the cursor.

**Step 1: Write tests**

Test that `SlashMenu(items)` renders correct number of menu items. Test that menu items have correct labels.

**Step 2: Implement**

`SlashMenuItems` — predefined list:
- Text, Heading 1, Heading 2, Heading 3, Bullet List, Numbered List, Toggle List, Code, Divider

`SlashMenu` component:
- Absolute positioned div with class `slash-menu`
- Each item: icon + label, click dispatches `ChangeBlockType` and closes menu
- Keyboard navigation: up/down arrows, Enter to select, Escape to close
- Filter: typing after `/` filters menu items by name match
- Mobile: renders as bottom sheet (positioned at bottom of viewport)

**Step 3: Commit**

```bash
git commit -m "feat(notion): add slash command menu for block type selection"
```

---

### Task 14: Editor Page Assembly

**Files:**
- Create: `src/app/components/editor_page.go` + `editor_page_stub.go`
- Test: `test/e2e/notion/editor_test.go`

**Context:** Wire everything together. `EditorPage(pageID)` loads the page's blocks from store (or server), renders `PageHeader` + `BlockList`. BlockList renders each block in order using `RenderBlock`.

**Step 1: Write E2E tests**

Test full render pipeline: create workspace with a page and blocks, render `EditorPage`, verify HTML output contains expected content.

**Step 2: Implement**

`EditorPage(pageID string) *dom.Element`:
1. Dispatch `LoadPage` with blocks from store/server
2. Subscribe to editor state changes
3. Return `dom.Div(dom.Class("editor-page"), PageHeader(page), BlockList(editorState))`

`BlockList(es *models.EditorState) *dom.Element`:
- Iterate `es.BlockOrder`, render each block via `RenderBlock(es.Blocks[id])`
- Empty state: render placeholder "Press Enter to start writing..."

WASM version: subscribes to store `"editor"` key, re-renders BlockList on changes.

**Step 3: Commit**

```bash
git commit -m "feat(notion): assemble editor page with page header and block list"
```

---

### Task 15: SSR Integration

**Files:**
- Modify: `internal/ssr/renderer.go` (add initial state embedding)
- Create: `src/app/ssr_setup.go` (non-WASM — registers routes for SSR)
- Test: `test/e2e/notion/ssr_test.go`

**Context:** SSR needs to: (1) load workspace + page data from server, (2) register routes with the non-WASM router, (3) render to HTML, (4) embed initial state as a `<script>` tag so WASM can hydrate without refetching.

**Step 1: Write tests**

Test that SSR output contains: DOCTYPE, title, sidebar HTML with page tree, editor HTML with block content, the WASM bootstrap script, and an embedded `<script id="__GOLEM_STATE__">` tag with JSON state.

**Step 2: Implement**

SSR route setup (non-WASM):
```go
func SetupSSRRoutes(r *router.Router, ws *models.Workspace, loadBlocks func(string) []*models.Block) {
    r.AddRoute(&router.Route{
        Path: "/:pageId",
        Component: func(params map[string]string) *dom.Element {
            pageID := params["pageId"]
            blocks := loadBlocks(pageID)
            return EditorPage(pageID, ws, blocks) // SSR version takes data directly
        },
        Layout: func(child *dom.Element) *dom.Element {
            return AppLayout(SidebarSSR(ws, ""), child)
        },
    })
}
```

State embedding: after `renderToDocument`, inject `<script id="__GOLEM_STATE__" type="application/json">` with JSON-serialized workspace + blocks before the WASM bootstrap.

WASM hydration: on load, check for `__GOLEM_STATE__` script tag, parse JSON, initialize store with that data, call `dom.Hydrate` instead of `dom.Render`.

**Step 3: Commit**

```bash
git commit -m "feat(notion): add SSR with embedded state and WASM hydration"
```

---

### Task 16: Mobile Responsive Sidebar

**Files:**
- Modify: `src/app/components/sidebar.go` (add drawer toggle logic)
- Modify: `src/app/styles/styles.go` (mobile media queries already added in Task 6)
- Test: `test/unit/components/sidebar_test.go` (add mobile tests)

**Context:** On mobile (< 768px), sidebar is hidden by default. A hamburger button in the top-left toggles the sidebar as a slide-over drawer with an overlay.

**Step 1: Write tests**

Test that sidebar renders a hamburger button. Test that the sidebar has the `sidebar-mobile` class variant. Test that the overlay element exists.

**Step 2: Implement**

Add to `Sidebar()`:
- A hamburger button (`☰`) visible only on mobile (CSS hides on desktop)
- WASM: clicking hamburger toggles a `sidebar-open` class on the sidebar
- An overlay div that covers the editor when sidebar is open
- Clicking overlay or navigating to a page closes the sidebar
- CSS transitions for smooth slide-in/out (already in Task 6 CSS)

**Step 3: Commit**

```bash
git commit -m "feat(notion): add mobile responsive sidebar with drawer toggle"
```

---

## Task Dependency Order

```
Task 1 (data model) ─┬─> Task 2 (workspace reducer) ─┬─> Task 5 (middleware) ──> Task 7 (layout/routing)
                      │                                │
                      └─> Task 3 (editor reducer) ─────┘
                      │
                      └─> Task 4 (server functions)

Task 6 (CSS) ──────────────────────────────────────────────> Task 7 (layout/routing)

Task 7 ─┬─> Task 8 (sidebar) ──> Task 16 (mobile sidebar)
        │
        └─> Task 9 (page header) ──> Task 14 (editor page assembly)
        │
        └─> Task 10 (basic blocks) ──> Task 11 (list/toggle/code) ──> Task 12 (interactions) ──> Task 13 (slash menu)

Task 14 (editor assembly) + Task 15 (SSR) are final integration tasks.
```

**Parallelizable groups:**
- Tasks 2 + 3 can run in parallel (both depend only on Task 1)
- Tasks 4 + 5 + 6 can run in parallel
- Tasks 8 + 9 + 10 can run in parallel (all depend on Task 7)
- Tasks 11 and 16 are independent of each other

---

## Summary

| Task | Description | Depends On |
|------|-------------|------------|
| 1 | Data model types | — |
| 2 | Workspace reducer | 1 |
| 3 | Editor reducer | 1 |
| 4 | Server-side persistence | 1 |
| 5 | Store middleware | 2, 3 |
| 6 | CSS design system | — |
| 7 | App layout & routing | 5, 6 |
| 8 | Sidebar components | 7 |
| 9 | Page header | 7 |
| 10 | Basic block types | 7 |
| 11 | List, toggle, code blocks | 10 |
| 12 | Block editing interactions | 10 |
| 13 | Slash command menu | 12 |
| 14 | Editor page assembly | 8, 9, 11, 12 |
| 15 | SSR integration | 14 |
| 16 | Mobile responsive sidebar | 8 |
