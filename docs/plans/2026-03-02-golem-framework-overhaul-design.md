# Golem Framework Overhaul Design

**Date**: 2026-03-02
**Status**: Approved
**Scope**: Full framework overhaul with Next.js-equivalent core features

## Overview

Golem is a Go-to-WebAssembly frontend framework. The current codebase has working primitives (dom, state, router, css) but the build pipeline, dev server, and advanced features are incomplete. This design covers fixing the foundation and adding core features equivalent to Next.js's App Router architecture, adapted for Go/WASM.

## Architectural Spine

Golem is a **two-phase compiler**:
1. **Phase 1 (Golem CLI)**: Scan file conventions in `src/app/`, generate Go source code (route registrations, layout wrappers, server stubs, error boundaries)
2. **Phase 2 (Go compiler)**: Compile generated + user code to WASM (`GOOS=js GOARCH=wasm`)

The **dual rendering model** uses Go build tags:
- `//go:build js && wasm` — client code that manipulates the browser DOM
- `//go:build !js || !wasm` — server code that produces HTML strings

The same component code compiles for both targets.

## Phase 1: Fix What's Broken

### 1.1 Build Pipeline (`internal/build/builder.go`)

**Problem**: `createWasmMainFile()` merges two Go files by stripping `package main`, breaking imports and build tags.

**Fix**: Compile `./src/app/` directly as a Go package:
```
go build -o .golem/dev/app.wasm ./src/app/
```

No file merging. The app's `main.go` imports its own components and server packages.

### 1.2 Dev Server (`internal/dev/server.go`)

**Problem**: `watchFiles()` is a stub. `handleWebSocket()` accepts connections but never sends reload messages.

**Fix**:
- Implement file polling (check mtimes every 500ms on `src/` directory tree)
- On change detected: recompile WASM, broadcast `"reload"` via WebSocket to all connected clients
- Capture `go build` stderr and send compile errors via WebSocket for error overlay
- Track compilation time and display in console

### 1.3 Module Structure

**Problem**: `golem-test-app/` has its own `go.mod` with no `replace` directive for local development.

**Fix**: Example apps use `replace` directives to point at the local framework. The main repo's `src/` directory is the canonical example app.

## Phase 2: Core Features

### 2.1 File-Based Routing with Code Generation

**Convention** (scanned at build time from `src/app/`):

| File | Exported Function | Purpose |
|------|-------------------|---------|
| `page.go` | `Page(params map[string]string) *dom.Element` | Route UI |
| `layout.go` | `Layout(children *dom.Element) *dom.Element` | Persistent wrapper |
| `template.go` | `Template(children *dom.Element) *dom.Element` | Re-mounting wrapper |
| `loading.go` | `Loading() *dom.Element` | Loading fallback |
| `error.go` | `Error(err error) *dom.Element` | Error boundary fallback |
| `notfound.go` | `NotFound() *dom.Element` | 404 UI |

**Directory patterns**:
- `blog/[slug]/page.go` → `/blog/:slug` (dynamic segment)
- `docs/[...path]/page.go` → `/docs/*` (catch-all)
- `api/[[...path]]/page.go` → `/api` and `/api/*` (optional catch-all)
- `(marketing)/about/page.go` → `/about` (route group, no URL impact)
- `@sidebar/page.go` → named parallel slot
- `(..)photo/[id]/page.go` → intercepting route

**Build-time output**: A `routes_gen.go` file is generated that registers all routes with the router, wiring layouts, error boundaries, loading states, parallel slots, and intercepting routes.

### 2.2 Layout System

- Layouts wrap child routes and persist across navigation (no re-render)
- Layouts nest: root layout → section layout → page
- Templates (`template.go`) re-mount on every navigation (fresh state)
- Rendering hierarchy: `layout → template → error boundary → loading → page`
- Root layout is required (`src/app/layout.go`)

### 2.3 Error Boundaries

- Use Go's `recover()` to catch panics during component rendering
- Error boundaries render fallback UI without crashing the whole app
- Errors bubble up to the nearest parent `error.go` boundary
- `error.go` at the same level as `layout.go` does NOT catch errors from that layout (boundary is a child of the layout)
- Root error boundary: `src/app/error.go`
- Global error boundary: `src/app/global-error.go` (catches root layout errors)

### 2.4 SSR (Server-Side Rendering)

**Dual rendering via build tags**:

The `_stub.go` files become full server-side renderers:

- `dom/element_server.go` (`!js || !wasm`): `RenderToHTML(*Element) string` traverses the element tree and produces HTML strings
- `dom/hydrate.go` (`js && wasm`): `Hydrate(*Element, selector string)` walks existing server-rendered DOM and attaches event handlers instead of replacing innerHTML
- `state/reactive_server.go`: Observables return initial values (no reactivity needed server-side)

**Request flow**:
1. Browser requests `/blog/hello`
2. Go server matches route, runs `Page(params)` natively (not WASM)
3. Wraps in layout chain, produces full HTML
4. Sends HTML with embedded `<script>` to load WASM
5. WASM loads, hydrates existing DOM (attaches event handlers)
6. Page becomes interactive

### 2.5 Static Generation

- At build time, run each page through the server renderer
- Output `.html` files to the build directory
- Pages with dynamic segments need `GenerateStaticParams()` to enumerate paths
- Static HTML served directly (no server rendering needed)

```go
// src/app/blog/[slug]/page.go
func GenerateStaticParams() []map[string]string {
    return []map[string]string{
        {"slug": "hello-world"},
        {"slug": "getting-started"},
    }
}
```

### 2.6 ISR (Incremental Static Regeneration)

```go
func PageConfig() golem.PageOptions {
    return golem.PageOptions{
        Revalidate: 60, // regenerate after 60 seconds
    }
}
```

- Server serves cached HTML for initial request
- After `Revalidate` seconds, next request triggers background regeneration
- Stale content served while regenerating (stale-while-revalidate pattern)
- Fresh content available for subsequent requests
- On-demand: `golem.RevalidatePath("/blog/hello")` and `golem.RevalidateTag("posts")`

### 2.7 Data Caching

```go
func GetPosts() ([]Post, error) {
    return golem.Cached("posts", golem.CacheOptions{
        Revalidate: 300,
        Tags:       []string{"blog-posts"},
    }, func() ([]Post, error) {
        // actual data fetch
    })
}
```

- In-memory cache with TTL per entry
- Tag-based invalidation
- Optional disk persistence for cache survival across restarts

### 2.8 True HMR (Hot Module Replacement)

**Architecture**: Shell WASM + swappable page WASM modules.

- **Shell WASM** (persistent): Router, layout system, state store, CSS engine. Stays loaded.
- **Page WASM modules** (swappable): Each route compiles to a separate WASM module.

**On file change**:
1. Detect which page was modified
2. Recompile only that page module (~100ms)
3. Send new module URL via WebSocket
4. JS bridge loads new module via `WebAssembly.instantiateStreaming`
5. Shell re-renders affected route with new module
6. State preserved (lives in shell)

**Go 1.24+ `//go:wasmexport`**: Each page exports `RenderPage(params)` as a WASM export.

**Fallback**: For changes that touch the shell (router, layouts, state), full reload with state serialization to sessionStorage.

### 2.9 Parallel Routes

**Convention**: `@slot` directories inside a route segment.

```
src/app/
  layout.go              → receives slots map
  @sidebar/
    page.go              → renders in "sidebar" slot
    loading.go           → independent loading state
  @content/
    page.go              → renders in "content" slot
    error.go             → independent error boundary
```

- Layout signature: `func Layout(slots map[string]*dom.Element) *dom.Element`
- Each slot renders independently with its own loading/error states
- Slots can navigate independently
- Default slot name is "children" for non-parallel route content

### 2.10 Intercepting Routes

**Convention**: Parenthesized dot prefixes.

| Pattern | Meaning |
|---------|---------|
| `(.)route` | Intercept same level |
| `(..)route` | Intercept one level up |
| `(...)route` | Intercept from root |

- Client-side navigation: render intercepted version (e.g., in modal/overlay)
- Direct URL access: render full page version
- Build-time scanner recognizes patterns and generates appropriate routing logic

### 2.11 Component Lifecycle

- `OnMount(fn func())` — runs after component is added to DOM
- `OnUnmount(fn func())` — runs before component is removed
- `OnUpdate(fn func())` — runs after re-render
- Cleanup functions returned by OnMount run on unmount
- Integrates with Observable reactive system

### 2.12 Enhanced Server Functions

- Build-time discovery of functions in `src/server/`
- Auto-generate type-safe client stubs callable from WASM
- HTTP/JSON transport via existing gRPC bridge
- Server functions support `"use server"` equivalent via package convention

### 2.13 Middleware

**Server middleware** (`src/middleware.go`):
```go
func Middleware(req *golem.Request, next func(*golem.Request) *golem.Response) *golem.Response {
    // auth check, redirect, modify headers, etc.
    return next(req)
}

func MiddlewareConfig() golem.MiddlewareConfig {
    return golem.MiddlewareConfig{
        Matcher: []string{"/dashboard/:path*", "/api/:path*"},
    }
}
```

**Client middleware** (router guards): Already exists via `BeforeEach` guards.

## Phase 3: Test Suite

### Structure

```
test/
├── unit/                              # go test — no WASM, no browser
│   ├── dom/
│   │   ├── element_test.go            # Element creation, attributes, children
│   │   ├── render_html_test.go        # RenderToHTML produces valid HTML
│   │   ├── hydration_test.go          # Hydrate logic correctness
│   │   └── vdom_diff_test.go          # Virtual DOM diffing algorithm
│   ├── state/
│   │   ├── observable_test.go         # Observable get/set/subscribe
│   │   ├── store_test.go             # Store dispatch/reduce/middleware
│   │   ├── computed_test.go           # Computed values
│   │   └── persistence_test.go        # LocalStorage serialization
│   ├── router/
│   │   ├── matching_test.go           # Route matching, param extraction
│   │   ├── guards_test.go            # Route guards
│   │   ├── parallel_test.go           # Parallel route resolution
│   │   └── intercepting_test.go       # Intercepting route logic
│   ├── css/
│   │   ├── stylesheet_test.go         # CSS generation
│   │   ├── theme_test.go             # Theme system
│   │   └── injection_test.go          # Style injection
│   ├── codegen/
│   │   ├── scanner_test.go            # File convention scanning
│   │   ├── route_gen_test.go          # Route code generation
│   │   └── stub_gen_test.go           # Server function stub generation
│   ├── cache/
│   │   ├── memory_cache_test.go       # In-memory cache TTL
│   │   ├── tag_invalidation_test.go   # Tag-based invalidation
│   │   └── revalidation_test.go       # Revalidation logic
│   └── functions/
│       ├── registry_test.go           # Function registration
│       ├── discovery_test.go          # Auto-discovery from directory
│       └── invocation_test.go         # Function calling
│
├── integration/                       # go test — compile WASM + test server
│   ├── build_pipeline_test.go         # Full build produces valid output
│   ├── ssr_test.go                    # Component → HTML → serve → verify
│   ├── static_gen_test.go            # Build-time static generation
│   ├── isr_test.go                    # Revalidation after TTL
│   ├── routing_codegen_test.go        # Generated routes match conventions
│   ├── server_functions_test.go       # gRPC function calling end-to-end
│   ├── middleware_test.go             # Server middleware pipeline
│   ├── parallel_routes_test.go        # Multi-slot rendering
│   ├── intercepting_routes_test.go    # Client vs direct URL behavior
│   └── error_handling_test.go         # Error boundaries catch and recover
│
├── development/                       # go test — dev server behavior
│   ├── hot_reload_test.go            # File change → rebuild → reload
│   ├── hmr_test.go                   # Module swap without full reload
│   ├── error_overlay_test.go          # Compile errors shown in browser
│   ├── dev_server_test.go            # Server starts, serves, handles API
│   └── file_watching_test.go          # Watcher detects changes
│
├── e2e/                               # Playwright browser tests
│   ├── basic_rendering_test.go        # App loads, renders initial UI
│   ├── ssr_hydration_test.go          # Server HTML → WASM hydrate → interactive
│   ├── navigation_test.go            # Route changes update DOM
│   ├── dynamic_routes_test.go         # Parameter routes work
│   ├── layout_persistence_test.go     # Layouts survive navigation
│   ├── error_boundary_test.go         # Errors show fallback UI
│   ├── state_reactivity_test.go       # Click → state → UI update
│   ├── server_functions_test.go       # Call server fn, get response
│   ├── hmr_browser_test.go           # File change → state preserved
│   ├── isr_browser_test.go           # Stale → revalidate → fresh
│   ├── parallel_rendering_test.go     # Independent slot loading/error
│   ├── intercepting_modal_test.go     # Client nav → modal, direct → page
│   ├── middleware_redirect_test.go    # Middleware blocks/redirects
│   ├── static_export_test.go          # Static HTML loads without server
│   └── form_handling_test.go          # Form → server action → response
│
└── testutil/
    ├── server.go                      # Start/stop test dev server
    ├── wasm.go                       # WASM compilation helpers
    ├── browser.go                    # Playwright browser helpers
    ├── fixture.go                    # Test app scaffolding
    └── assertion.go                  # Custom test assertions
```

### Test Approach

- **Unit tests**: Standard `go test`. Run without WASM or browser. Test pure logic.
- **Integration tests**: Compile real WASM, start real server, verify behavior programmatically.
- **Development tests**: Start dev server, modify files, verify rebuild/reload behavior.
- **E2E tests**: Playwright browser automation. Full user interaction flows.
- **All tests run in CI** via GitHub Actions.

## Implementation Order

1. Fix build pipeline (WASM compilation works)
2. Fix dev server (file watching + WebSocket reload)
3. Server-side HTML rendering (`RenderToHTML`)
4. File-based routing code generation
5. Layout system
6. Error boundaries
7. Hydration
8. Component lifecycle
9. Server functions (enhanced)
10. Static generation
11. ISR + caching
12. Middleware
13. HMR (module splitting)
14. Parallel routes
15. Intercepting routes
16. Test suite (built incrementally alongside features)

## Out of Scope

- Rich text editing (contenteditable in WASM)
- Image optimization (not relevant for WASM framework)
- Font optimization (CSS handles this)
- i18n (future addition)
- Database adapters (use server functions)
