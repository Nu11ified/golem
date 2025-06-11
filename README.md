# Go React Meta-Framework

## 1. What is this project? (Goals)

This is a proof-of-concept meta-framework for building high-performance, **Go-powered Server Driven UI (SDUI)** React applications. The server (Go) is in full control of the UI, orchestrating which components and layouts are rendered, and delivering pre-rendered HTML to the browser for hydration. The goals are:
- Provide a modern SSR React stack with Go as the public-facing server and orchestrator.
- Enable true server-driven UI: the server determines the UI structure and content for every route.
- Hide all build and hydration complexity from the user—users only write React components, layouts, and static assets in `user-app/`.
- All SSR/hydration logic (including the hydration entry point and import map) lives in `node-renderer/` for clean separation.
- Enable fast development and production workflows with a single command for each.
- Support Next.js-style layouts, metadata/SEO, and file-based routing.
- Allow for future extensibility (e.g., HMR, data fetching, etc.).

---

## 2. Architecture Diagram: Go-Powered Server Driven UI

```mermaid
graph TD
  subgraph Client
    A["Browser"]
  end

  subgraph Server
    B["Go Orchestrator (:8080)"]
    C["Node Renderer (:3001)"]
    F["Go Server Functions\n(user-app/server/go)"]
    G["TS Server Functions\n(user-app/server/ts)"]
  end

  subgraph UserApp
    D["pages/*.tsx, layout.tsx, components/*"]
  end

  subgraph NodeRenderer
    E["hydrate.tsx, importMap.generated.js"]
  end

  %% Main SSR/SDUI flow
  A -- "HTTP Request" --> B
  B -- "POST /render {component, layout, props}" --> C
  C -- "SSR React HTML + metadata" --> B
  B -- "HTML + <script>client.js</script> + <script>window.__SSR_PROPS__</script>" --> A
  C -- "Serves /client.js (hydration bundle)" --> A

  %% File-based routing and hydration
  D -- "imported by" --> C
  E -- "hydration logic" --> A

  %% Server function flows
  B -- "API: /api/go/{fn}" --> F
  B -- "API: /api/ts/{fn}" --> G
  F -- "Dynamic Plugin Load" --> B
  G -- "Node Runner" --> B

  %% UI orchestration
  B -- "Controls UI structure (Server Driven)" --> C
=======
    A["Browser"] -- "HTTP Request" --> B["Go Orchestrator (:8080)"]
    B -- "POST /render {component, layout, props}" --> C["Node Renderer (:3001)"]
    C -- "SSR React HTML + metadata" --> B
    B -- "HTML + <script>client.js</script> + <script>window.__SSR_PROPS__</script>" --> A
    C -- "Serves /client.js (hydration bundle)" --> A
    D["user-app/pages/*.tsx, layout.tsx, components/*"] -- "imported by" --> C
    E["node-renderer/hydrate.tsx, importMap.generated.js"] -- "hydration logic" --> A
    B -- "Controls UI structure (Server Driven)" --> C
    B -- "API: /api/go/{fn}" --> F[Go Server Functions (plugins in user-app/server/go)]
    B -- "API: /api/ts/{fn}" --> G[TS Server Functions (user-app/server/ts)]
    F -- "Dynamic Plugin Load" --> B
    G -- "Node Runner" --> B
```

## 2a. Server Function System

This framework supports polyglot server functions:
- **Go server functions**: Place `.go` files in `user-app/server/go`. These are built as Go plugins (`.so` files) and hot-reloaded automatically in development. The orchestrator loads and executes them dynamically for `/api/go/{functionName}` requests.
- **TypeScript server functions**: Place `.ts` files in `user-app/server/ts`. These are loaded and executed dynamically by a Node.js runner for `/api/ts/{functionName}` requests. No build step is needed for TypeScript functions.
- **Automation**: The dev workflow automatically watches and rebuilds Go plugins, and all server function changes are picked up instantly without restarting the orchestrator.

### Example Usage
- Add a Go function: `user-app/server/go/hello.go` (exports a `Handler` function)
- Add a TypeScript function: `user-app/server/ts/hello.ts` (exports a default async function)
- Call them from the frontend or via `/api/go/hello` and `/api/ts/hello`

## 3. How the Server-Driven UI Schema Works

This framework implements a **platform-agnostic, server-driven UI** using a JSON Schema-based approach:

- The Go server exposes two key endpoints:
  - **`/ui-schema`**: Serves the platform-agnostic UI schema as JSON Schema. This schema defines all possible UI primitives (Screen, Text, Button, Input, List, etc.) and their properties.
  - **`/ui`**: Serves a sample UI data object that matches the schema. This represents the actual UI to be rendered for a given route or user.

- **Any client** (web, mobile, desktop, etc.) can:
  1. **Fetch the schema** from `/ui-schema` and generate TypeScript (or other language) types for type-safe UI rendering.
  2. **Fetch the UI data** from `/ui` (or a similar endpoint for real apps) and render the UI using their platform's native components, following the schema.

- This enables you to:
  - Update UI and flows from the server without redeploying clients.
  - Share a single source of truth for UI structure across all platforms.
  - Achieve true server-driven UI for any device that can parse JSON and render UI.

**Example Flow:**
1. The client fetches `/ui-schema` and generates types.
2. The client fetches `/ui` and receives a UI description like:
   ```json
   {
     "type": "Screen",
     "props": { "title": "Welcome" },
     "children": [
       { "type": "Text", "props": { "value": "Hello, user!" } },
       { "type": "Button", "props": { "label": "Click me", "action": "doSomething" } }
     ]
   }
   ```
3. The client renders the UI using its own primitives (React, React Native, etc.), using the generated types for safety.

## 4. Who Uses Server-Driven UI? (Industry Adoption & References)

Server-Driven UI (SDUI) is used by many leading tech companies to enable rapid iteration, cross-platform consistency, and dynamic feature rollout. Here are some notable examples:

### Lyft
- **Why:** To manage business complexity, increase release velocity, and support highly configurable, market-specific experiences in their Bikes & Scooters and Rideshare apps.
- **How:** Uses a "Backend for Frontend" (BFF) microservice to deliver UI schemas and actions to clients, with both declarative and semantic components.
- **Reference:** [The Journey to Server Driven UI At Lyft Bikes and Scooters (Lyft Engineering Blog)](https://eng.lyft.com/the-journey-to-server-driven-ui-at-lyft-bikes-and-scooters-c19264a0378e)

### Airbnb
- **Why:** To enable A/B testing, rapid UI changes, and consistent experiences across platforms without requiring app store releases.
- **How:** Uses SDUI for parts of their mobile app, sending UI schemas from the server to the client for rendering.
- **Reference:** [Judo: What is Server-Driven UI?](https://www.judo.app/blog/server-driven-ui)

### Shopify
- **Why:** For dynamic onboarding and flows in their mobile apps, allowing instant updates and experiments.
- **How:** Uses SDUI to deliver onboarding screens and other flows as server-defined schemas.
- **Reference:** [Unlocking the Power of Server-Driven UI (Medium)](https://medium.com/@dimakoua/unlocking-the-power-of-server-driven-ui-building-dynamic-configurable-apps-16a9f5bdf95a)

### Judo
- **Why:** Provides a commercial SDUI solution for mobile teams, enabling real-time UI updates and feature rollouts.
- **How:** Offers an SDK and platform for building and rendering server-driven UIs in native apps.
- **Reference:** [Judo: What is Server-Driven UI?](https://www.judo.app/blog/server-driven-ui)

---

**Why do these companies use SDUI?**
- **Faster iteration:** UI changes can be made server-side and instantly reflected in all clients.
- **A/B testing:** Easily test different UI layouts or flows without new app releases.
- **Consistency:** Ensure all platforms show the same UI, reducing fragmentation.
- **Feature rollout:** Roll out new features or UI changes to all users at once, or segment by user group.
- **Reduced release cycle pain:** No need to wait for app store approvals for UI changes.

**References:**
- [The Journey to Server Driven UI At Lyft Bikes and Scooters (Lyft Engineering Blog)](https://eng.lyft.com/the-journey-to-server-driven-ui-at-lyft-bikes-and-scooters-c19264a0378e)
- [Judo: What is Server-Driven UI?](https://www.judo.app/blog/server-driven-ui)
- [Unlocking the Power of Server-Driven UI (Medium)](https://medium.com/@dimakoua/unlocking-the-power-of-server-driven-ui-building-dynamic-configurable-apps-16a9f5bdf95a)

## 5. Generating TypeScript Types from the UI Schema

To ensure type safety and a great developer experience, you can generate TypeScript types from the platform-agnostic UI schema using [quicktype](https://quicktype.io).

### One-time Type Generation

Run:
```sh
pnpm --filter user-app run generate:types
```
This will generate `user-app/ui-schema.d.ts` from `user-app/ui-schema.json`.

### Watch for Changes

To automatically regenerate types whenever you edit `ui-schema.json`, run:
```sh
pnpm --filter user-app run watch:types
```

### Sample TypeScript Usage

After generating, you can use the types in your app:

```typescript
import { Screen } from './ui-schema';

function renderUI(screen: Screen) {
  return (
    <div>
      <h1>{screen.props.title}</h1>
      {screen.children?.map((child, idx) => {
        switch (child.type) {
          case 'Text':
            return <p key={idx}>{child.props.value}</p>;
          case 'Button':
            return <button key={idx} onClick={() => handleAction(child.props.action)}>{child.props.label}</button>;
          case 'Input':
            return <input key={idx} name={child.props.name} type={child.props.inputType} placeholder={child.props.label} />;
          case 'List':
            return (
              <ul key={idx}>
                {child.props.items.map((item: string, i: number) => <li key={i}>{item}</li>)}
              </ul>
            );
          default:
            return null;
        }
      })}
    </div>
  );
}

function handleAction(action: string) {
  alert(`Action: ${action}`);
}
```

### Troubleshooting

If you see an error like `command not found: quicktype`, make sure you are using the scripts with `npx` (as in the provided scripts). If you have network issues, you can install quicktype globally:

```sh
npm install -g quicktype
```

and then run:

```sh
quicktype -s schema ui-schema.json -o ui-schema.d.ts
```

But for most users, the provided scripts using `npx quicktype` should work out of the box.

---

## 6. Getting Started

### Prerequisites
- [Go](https://golang.org/) (v1.20+)
- [Node.js](https://nodejs.org/) (v18+ recommended)
- [pnpm](https://pnpm.io/) (v8+ recommended)

### Install dependencies and build everything
```bash
pnpm install
```

### Development mode (auto-rebuild, live reload)
```bash
pnpm dev
```
- Starts both the Go orchestrator and Node renderer (with client bundle watcher).
- Edit files in `user-app/pages/`, `user-app/components/`, or `user-app/public/` and refresh the browser to see changes and interactivity.
- **Automatic import map generation:** All pages and layouts are auto-discovered for hydration.

### Production build
```bash
pnpm build
```
- Builds the Go binary and the static client bundle.

### Production run
```bash
pnpm start
```
- Runs both the Go orchestrator and Node renderer using the production builds.

### Visit your app
Open [http://localhost:8080](http://localhost:8080) in your browser.

---

## 7. Features

### 🗂️ File-Based Routing
- Any `.tsx` file in `user-app/pages/` becomes a route (including nested folders).
- Example: `pages/about.tsx` → `/about`, `pages/blog/post.tsx` → `/blog/post`.

### 🧩 Next.js-Style Layouts
- Add a `layout.tsx` in any folder in `pages/` to wrap all pages in that folder (and subfolders).
- Example: `pages/layout.tsx` is the global layout; `pages/blog/layout.tsx` wraps all `/blog/*` pages.
- Layouts receive `children` as a prop.

### 📝 Metadata & SEO
- Export a `metadata` object from any page or layout:
  ```tsx
  // pages/layout.tsx
  export const metadata = {
    title: 'My Custom Site',
    description: 'A modern Go/React meta-framework demo',
    favicon: '/favicon.ico',
  };
  ```
- Metadata is merged (layout takes precedence) and injected into the HTML `<head>` (title, description, favicon, etc.).

### ⚡ Automatic Import Map Generation
- All pages and layouts are auto-discovered and included in the client bundle for hydration.
- No need to manually update import maps—just add files to `pages/`.
- **All hydration logic and the import map now live in `node-renderer/` for separation of concerns.**

### 🖼️ Static Assets
- Place static files in `user-app/public/` (e.g., `favicon.ico`, images, etc.).
- Access them at `/filename.ext` (e.g., `/favicon.ico`).

### 🧩 Components
- Place reusable components in `user-app/components/` and import them into your pages/layouts.

---

## 8. Example: Layout, Metadata, and Component Usage

**user-app/pages/layout.tsx**
```tsx
import React from 'react';

export const metadata = {
  title: 'My Custom Site',
  description: 'A modern Go/React meta-framework demo',
  favicon: '/favicon.ico',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ minHeight: '100vh', background: '#222', color: '#fff' }}>
      <header style={{ padding: '1rem', background: '#61dafb', color: '#222', fontWeight: 'bold', fontSize: '1.5rem' }}>
        🚀 My Custom Site
      </header>
      <main style={{ padding: '2rem' }}>{children}</main>
    </div>
  );
}
```

**user-app/components/InteractiveButton.tsx**
```tsx
import React, { useState } from 'react';

export default function InteractiveButton() {
  const [clicked, setClicked] = useState(false);
  return (
    <button
      style={{ marginTop: '2rem', padding: '0.5rem 1.5rem', fontSize: '1.1rem', borderRadius: '6px', border: 'none', background: '#61DAFB', color: '#222', cursor: 'pointer' }}
      onClick={() => setClicked(c => !c)}
    >
      {clicked ? 'You clicked me!' : 'Click me!'}
    </button>
  );
}
```

**user-app/pages/index.tsx**
```tsx
import React from 'react';
import InteractiveButton from '../components/InteractiveButton';

export default function HomePage({ frameworkName }: { frameworkName: string }) {
  return (
    <div style={{ fontFamily: 'sans-serif', padding: '2rem', border: '2px solid #61DAFB', borderRadius: '8px', maxWidth: '600px', margin: '2rem auto', textAlign: 'center' }}>
      <h1>Welcome to the Future!</h1>
      <p>This React component was server-side rendered by a <strong>{frameworkName}</strong> stack.</p>
      <InteractiveButton />
    </div>
  );
}
```

---

## 9. Current Limitations
- **No true HMR (Hot Module Replacement):** The client bundle is rebuilt and you must refresh the browser to see changes.
- **No dynamic data fetching conventions:** Props are static in the orchestrator; no API/data layer yet.
- **No error overlays or dev tools.**
- **No code splitting or asset optimization.**
- **No authentication/session support.**
- **No TypeScript type-checking in the build pipeline.**
- **No tests or CI/CD pipeline yet.**

---

## 10. How to Contribute

1. **Fork this repo and clone your fork.**
2. **Create a new branch for your feature or fix.**
3. **Make your changes.**
    - For framework changes: edit files in `go-orchestrator/` or `node-renderer/`.
    - For user-facing app: edit files in `user-app/pages/`, `user-app/components/`, or `user-app/public/`.
4. **Test your changes:**
    - Run `pnpm dev` and verify everything works as expected.
5. **Push your branch and open a pull request.**
6. **Describe your changes and why they're useful.**

---

## 11. Project Structure

```
my-go-framework/
├── bin/                  # Go binary output
├── go-orchestrator/      # Go server (public-facing)
├── node-renderer/        # Node.js SSR, hydration, and all build logic
│   ├── hydrate.tsx       # Hydration entry point (internal)
│   ├── importMap.generated.js # Auto-generated import map (internal)
│   └── ...               # Other SSR/build scripts
├── user-app/             # User React app (just write pages, components, and static assets!)
│   ├── pages/            # File-based routing & layouts
│   ├── components/       # Reusable React components
│   └── public/           # Static assets (favicon, images, etc.)
├── package.json          # Root scripts/workspaces
├── pnpm-workspace.yaml   # pnpm monorepo config
└── .gitignore
```

---

## 12. FAQ

**Q: Where do I write my app code?**
- Only in `user-app/pages/`, `user-app/components/`, or `user-app/public/`. Everything else is handled for you.

**Q: How do I add a new page?**
- Add a new `.tsx` file to `user-app/pages/` (e.g., `about.tsx`). It will be routed as `/about`.

**Q: How do I add a layout?**
- Add a `layout.tsx` file to any folder in `pages/`.

**Q: How do I add metadata/SEO?**
- Export a `metadata` object from your page or layout.

**Q: How do I add static assets?**
- Place them in `user-app/public/` and reference them by URL (e.g., `/logo.png`).

**Q: How do I add reusable components?**
- Place them in `user-app/components/` and import them into your pages/layouts.

**Q: How do I add dependencies?**
- Use `pnpm add <package> -F user-app` for user code, or `-F node-renderer` for framework code.

---

## 13. License

Mozilla Public License 2.0