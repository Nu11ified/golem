# Notion Clone Design

## Goal

Build a Notion-style workspace application using the Golem framework, showcasing SSR with hydration, reactive state management, server function RPC, and file-based routing. The app features a workspace shell with a page tree sidebar and a block-based editor supporting text, headings, lists, toggles, code blocks, and dividers.

## Architecture

Document-centric Redux-style Store with two slices ("workspace" for the page tree, "editor" for active page blocks). Middleware handles localStorage persistence (immediate) and server sync (debounced). SSR renders the initial page server-side with embedded state; WASM hydrates for interactivity.

## Data Model

### Page
```go
type Page struct {
    ID        string   `json:"id"`
    Title     string   `json:"title"`
    ParentID  string   `json:"parentId"`  // "" for root pages
    ChildIDs  []string `json:"childIds"`  // ordered child page IDs
    Icon      string   `json:"icon"`      // emoji icon
    CreatedAt int64    `json:"createdAt"`
    UpdatedAt int64    `json:"updatedAt"`
}
```

### Block
```go
type Block struct {
    ID       string                 `json:"id"`
    Type     string                 `json:"type"`     // "text", "h1", "h2", "h3", "bullet", "numbered", "toggle", "code", "divider"
    Content  string                 `json:"content"`  // text content
    Children []string               `json:"children"` // child block IDs (for toggle blocks)
    Props    map[string]interface{} `json:"props"`    // language for code blocks, etc.
    PageID   string                 `json:"pageId"`
}
```

### State Trees
```go
type Workspace struct {
    Pages       map[string]*Page  `json:"pages"`
    RootPageIDs []string          `json:"rootPageIds"`
}

type EditorState struct {
    ActivePageID string            `json:"activePageId"`
    Blocks       map[string]*Block `json:"blocks"`
    BlockOrder   []string          `json:"blockOrder"`
    FocusBlockID string            `json:"focusBlockId"`
}
```

Pages form a tree via ParentID/ChildIDs. Blocks are flat within a page (ordered by BlockOrder), except toggle blocks which have Children. This avoids deeply nested structures.

## Store Architecture

### Workspace Reducer Actions
- `CREATE_PAGE` — create page (optionally under a parent)
- `DELETE_PAGE` — remove page and reparent/delete children
- `RENAME_PAGE` — update page title
- `MOVE_PAGE` — reparent a page
- `SET_PAGE_ICON` — update emoji icon
- `REORDER_PAGES` — reorder children within a parent

### Editor Reducer Actions
- `LOAD_PAGE` — load blocks for a page
- `ADD_BLOCK` — insert block at position
- `DELETE_BLOCK` — remove a block
- `UPDATE_BLOCK` — update block content or type
- `MOVE_BLOCK` — reorder blocks
- `TOGGLE_BLOCK` — expand/collapse toggle block
- `CHANGE_BLOCK_TYPE` — convert block type (e.g., text to h1)

### Middleware Stack
1. **localStorage** — persists on every dispatch (immediate)
2. **Server sync** — debounced 500ms writes to server functions
3. **Logger** — dev-only action logging

## Component Tree

```
App (Layout)
+-- Sidebar (parallel slot)
|   +-- WorkspaceHeader ("My Workspace" + new page button)
|   +-- SearchBar (filters page tree)
|   +-- PageTree (recursive, collapsible)
|       +-- PageTreeItem (icon + title, indent by depth)
+-- EditorArea (main content)
    +-- PageHeader (icon picker + editable title)
    +-- BlockList
        +-- BlockRenderer (per block type)
            +-- TextBlock (contenteditable paragraph)
            +-- HeadingBlock (H1/H2/H3)
            +-- BulletListBlock
            +-- NumberedListBlock
            +-- ToggleBlock (collapsible with children)
            +-- CodeBlock (monospace, language label)
            +-- DividerBlock (horizontal rule)
```

## Routing

```
/              -> redirect to last opened page or first page
/[pageId]      -> page editor view
```

Layout wraps every route. Sidebar is a parallel slot. The [pageId] dynamic route renders the editor. SSR renders layout + sidebar + active page on first load.

## Block Editing Interactions

- Each block is a contenteditable div
- Enter: create new block below
- Backspace on empty block: delete and focus previous
- `# ` / `## ` / `### ` at start: convert to heading
- `- ` or `* `: convert to bullet list
- `1. `: convert to numbered list
- ` ``` `: convert to code block
- `---`: convert to divider
- `/`: open block type picker (slash command menu)

## Server Functions

```go
GetWorkspace() -> Workspace
SaveWorkspace(workspace Workspace)
GetPageBlocks(pageId string) -> []Block
SavePageBlocks(pageId string, blocks []Block)
DeletePageData(pageId string)
```

Storage: JSON files on disk (data/workspace.json, data/blocks/<pageId>.json).

## Sync Strategy

1. **On load:** localStorage loads instantly and renders. Background fetch from server. If server data newer (by UpdatedAt), merge in.
2. **On dispatch:** localStorage saves immediately. Server sync debounces 500ms after last change.
3. **Conflicts:** Last-write-wins by UpdatedAt. Sufficient for single-user.

## SSR Data Flow

1. Server receives request for /[pageId]
2. Loads workspace + page blocks from disk
3. SSR renderer builds full component tree
4. Returns HTML with embedded initial state in a script tag
5. WASM hydrates, reads initial state, skips first server fetch

## Styling

Global CSS via css.InjectStyles(). Class-based styling. Notion aesthetic.

### CSS Variables
```
--bg-primary: #ffffff
--bg-sidebar: #f7f6f3
--bg-hover: #efefef
--text-primary: #37352f
--text-secondary: #9b9a97
--accent: #2eaadc
--border: #e9e9e7
--font-body: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif
--font-mono: "SFMono-Regular", Consolas, monospace
```

### Layout
- Sidebar: 240px, light gray
- Editor: centered, max-width 720px, generous padding
- Typography: system font, 16px base, 1.6 line-height
- Blocks: no borders, subtle hover highlight, blue left border on focus

### Mobile (< 768px)
- Sidebar collapses to slide-over drawer with hamburger toggle
- Touch targets min 44px height
- Swipe-to-close on drawer
- Slash menu renders as bottom sheet
- Reduced editor padding
