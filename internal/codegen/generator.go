package codegen

import (
	"fmt"
	"go/format"
	"path/filepath"
	"sort"
	"strings"
)

// routeImport tracks an import needed for a page/layout/error package.
type routeImport struct {
	Alias      string // import alias (e.g., "page", "about", "blog_slug")
	ImportPath string // full import path (e.g., "myapp/app/about")
}

// parallelSlotEntry carries code-generation info for a single parallel slot.
type parallelSlotEntry struct {
	PageAlias  string // import alias for the slot's page package
	ErrorAlias string // import alias for the slot's error package (same package)
	HasError   bool   // whether the slot has an error.go file
	HasLoading bool   // whether the slot has a loading.go file
}

// routeEntry represents a single route registration to generate.
type routeEntry struct {
	Path             string
	PageAlias        string // import alias for the page package
	LayoutAlias      string // import alias for the layout package (may differ from page)
	ErrorAlias       string // import alias for the error package
	HasLayout        bool
	HasError         bool
	IsIntercepting   bool
	InterceptTarget  string
	ParallelSlots   map[string]*parallelSlotEntry // slot name -> slot entry
}

// GenerateRoutes takes a scanned route tree and module name, returns generated Go source code.
func GenerateRoutes(tree *ScannedRoute, moduleName string) (string, error) {
	imports := make(map[string]routeImport) // keyed by import path
	var entries []routeEntry

	// Collect all routes and imports from the tree
	collectRoutes(tree, moduleName, imports, &entries, "")

	// Build the source code
	var b strings.Builder

	// Build tag and package declaration
	b.WriteString("//go:build js && wasm\n\n")
	b.WriteString("package main\n\n")

	// Imports
	b.WriteString("import (\n")
	b.WriteString("\t\"github.com/Nu11ified/golem/dom\"\n")
	b.WriteString("\t\"github.com/Nu11ified/golem/router\"\n")

	// Sort imports for deterministic output
	sortedImports := sortImports(imports)
	for _, imp := range sortedImports {
		b.WriteString(fmt.Sprintf("\t%s %q\n", imp.Alias, imp.ImportPath))
	}
	b.WriteString(")\n\n")

	// Suppress unused import warnings for dom package
	b.WriteString("// Ensure dom is referenced.\n")
	b.WriteString("var _ = dom.Div\n\n")

	// registerRoutes function
	b.WriteString("func registerRoutes(r *router.Router) {\n")

	// Track layout variables for parent chaining
	layoutVarCounter := 0

	for _, entry := range entries {
		if entry.HasLayout && hasChildRoutes(entry, entries) {
			// Declare a variable for the route so children can reference it as parent
			layoutVarCounter++
			varName := fmt.Sprintf("route%d", layoutVarCounter)
			b.WriteString(fmt.Sprintf("\t%s := &router.Route{\n", varName))
		} else {
			b.WriteString("\tr.AddRoute(&router.Route{\n")
		}

		b.WriteString(fmt.Sprintf("\t\tPath: %q,\n", entry.Path))

		// Component (only if there is a page)
		if entry.PageAlias != "" {
			b.WriteString(fmt.Sprintf("\t\tComponent: func(params map[string]string) *dom.Element {\n"))
			b.WriteString(fmt.Sprintf("\t\t\treturn %s.Page(params)\n", entry.PageAlias))
			b.WriteString("\t\t},\n")
		}

		// Layout
		if entry.HasLayout {
			b.WriteString(fmt.Sprintf("\t\tLayout: func(children *dom.Element) *dom.Element {\n"))
			b.WriteString(fmt.Sprintf("\t\t\treturn %s.Layout(children)\n", entry.LayoutAlias))
			b.WriteString("\t\t},\n")
		}

		// Error handler
		if entry.HasError {
			b.WriteString(fmt.Sprintf("\t\tErrorHandler: func(err error) *dom.Element {\n"))
			b.WriteString(fmt.Sprintf("\t\t\treturn %s.Error(err)\n", entry.ErrorAlias))
			b.WriteString("\t\t},\n")
		}

		// Intercepting route
		if entry.IsIntercepting {
			b.WriteString("\t\tIsIntercepting: true,\n")
			if entry.InterceptTarget != "" {
				b.WriteString(fmt.Sprintf("\t\tInterceptTarget: %q,\n", entry.InterceptTarget))
			}
		}

		// Parallel slots
		if len(entry.ParallelSlots) > 0 {
			b.WriteString("\t\tParallelSlots: map[string]*router.Route{\n")
			// Sort slot names for deterministic output
			slotNames := make([]string, 0, len(entry.ParallelSlots))
			for name := range entry.ParallelSlots {
				slotNames = append(slotNames, name)
			}
			sort.Strings(slotNames)
			for _, name := range slotNames {
				slot := entry.ParallelSlots[name]
				b.WriteString(fmt.Sprintf("\t\t\t%q: {\n", name))
				b.WriteString(fmt.Sprintf("\t\t\t\tPath: %q,\n", entry.Path))
				b.WriteString(fmt.Sprintf("\t\t\t\tComponent: func(params map[string]string) *dom.Element {\n"))
				b.WriteString(fmt.Sprintf("\t\t\t\t\treturn %s.Page(params)\n", slot.PageAlias))
				b.WriteString("\t\t\t\t},\n")
				if slot.HasError {
					b.WriteString(fmt.Sprintf("\t\t\t\tErrorHandler: func(err error) *dom.Element {\n"))
					b.WriteString(fmt.Sprintf("\t\t\t\t\treturn %s.Error(err)\n", slot.ErrorAlias))
					b.WriteString("\t\t\t\t},\n")
				}
				if slot.HasLoading {
					b.WriteString(fmt.Sprintf("\t\t\t\tLoadingHandler: func() *dom.Element {\n"))
					b.WriteString(fmt.Sprintf("\t\t\t\t\treturn %s.Loading()\n", slot.PageAlias))
					b.WriteString("\t\t\t\t},\n")
				}
				b.WriteString("\t\t\t},\n")
			}
			b.WriteString("\t\t},\n")
		}

		if entry.HasLayout && hasChildRoutes(entry, entries) {
			b.WriteString("\t}\n")
			b.WriteString(fmt.Sprintf("\tr.AddRoute(%s)\n\n", varName(layoutVarCounter)))
		} else {
			b.WriteString("\t})\n\n")
		}
	}

	b.WriteString("}\n")

	// Format with go/format
	source := b.String()
	formatted, err := format.Source([]byte(source))
	if err != nil {
		// Return unformatted source with the error for debugging
		return source, fmt.Errorf("generated code formatting failed: %w\n\nGenerated source:\n%s", err, source)
	}

	return string(formatted), nil
}

// varName returns the variable name for the given counter value.
func varName(n int) string {
	return fmt.Sprintf("route%d", n)
}

// hasChildRoutes checks if any entries have this entry's path as a prefix parent.
func hasChildRoutes(entry routeEntry, entries []routeEntry) bool {
	for _, other := range entries {
		if other.Path != entry.Path && strings.HasPrefix(other.Path, entry.Path) {
			return true
		}
	}
	return false
}

// collectRoutes recursively walks the scanned route tree, collecting imports and route entries.
func collectRoutes(route *ScannedRoute, moduleName string, imports map[string]routeImport, entries *[]routeEntry, inheritedLayoutAlias string) {
	currentLayoutAlias := inheritedLayoutAlias

	// Determine the import alias for this route's directory
	dirAlias := dirToAlias(route.DirPath)

	// If this route has a layout file, register its import and track the alias
	if route.LayoutFile != "" {
		layoutImportPath := dirToImportPath(route.DirPath, moduleName)
		alias := ensureUniqueAlias(dirAlias, layoutImportPath, imports)
		imports[layoutImportPath] = routeImport{Alias: alias, ImportPath: layoutImportPath}
		currentLayoutAlias = alias
	}

	// If this route has a page file, register the route
	if route.PageFile != "" {
		pageImportPath := dirToImportPath(route.DirPath, moduleName)
		alias := ""
		if existing, ok := imports[pageImportPath]; ok {
			alias = existing.Alias
		} else {
			alias = ensureUniqueAlias(dirAlias, pageImportPath, imports)
			imports[pageImportPath] = routeImport{Alias: alias, ImportPath: pageImportPath}
		}

		entry := routeEntry{
			Path:      route.Path,
			PageAlias: alias,
		}

		// Layout: use the current route's layout or inherited layout
		if route.LayoutFile != "" {
			entry.HasLayout = true
			entry.LayoutAlias = currentLayoutAlias
		} else if inheritedLayoutAlias != "" {
			entry.HasLayout = true
			entry.LayoutAlias = inheritedLayoutAlias
		}

		// Error handler
		if route.ErrorFile != "" {
			entry.HasError = true
			entry.ErrorAlias = alias // same package as page
		}

		// Intercepting route
		if route.InterceptPattern != "" {
			entry.IsIntercepting = true
			entry.InterceptTarget = route.Path
		}

		// Parallel slots
		if len(route.ParallelSlots) > 0 {
			entry.ParallelSlots = make(map[string]*parallelSlotEntry)
			for slotName, slotRoute := range route.ParallelSlots {
				if slotRoute.PageFile != "" {
					slotImportPath := dirToImportPath(slotRoute.DirPath, moduleName)
					slotAlias := ""
					if existing, ok := imports[slotImportPath]; ok {
						slotAlias = existing.Alias
					} else {
						slotAlias = ensureUniqueAlias(dirToAlias(slotRoute.DirPath), slotImportPath, imports)
						imports[slotImportPath] = routeImport{Alias: slotAlias, ImportPath: slotImportPath}
					}
					se := &parallelSlotEntry{
						PageAlias:  slotAlias,
						HasError:   slotRoute.ErrorFile != "",
						HasLoading: slotRoute.LoadingFile != "",
					}
					if se.HasError {
						se.ErrorAlias = slotAlias // same package as page
					}
					entry.ParallelSlots[slotName] = se
				}
			}
		}

		*entries = append(*entries, entry)
	} else if len(route.ParallelSlots) > 0 {
		// Handle parallel slots even without a page file (e.g., layout-only routes)
		entry := routeEntry{
			Path:          route.Path,
			ParallelSlots: make(map[string]*parallelSlotEntry),
		}

		// Only emit parallel slot routes if they have pages.
		hasSlotPages := false
		for slotName, slotRoute := range route.ParallelSlots {
			if slotRoute.PageFile != "" {
				slotImportPath := dirToImportPath(slotRoute.DirPath, moduleName)
				slotAlias := ""
				if existing, ok := imports[slotImportPath]; ok {
					slotAlias = existing.Alias
				} else {
					slotAlias = ensureUniqueAlias(dirToAlias(slotRoute.DirPath), slotImportPath, imports)
					imports[slotImportPath] = routeImport{Alias: slotAlias, ImportPath: slotImportPath}
				}
				se := &parallelSlotEntry{
					PageAlias:  slotAlias,
					HasError:   slotRoute.ErrorFile != "",
					HasLoading: slotRoute.LoadingFile != "",
				}
				if se.HasError {
					se.ErrorAlias = slotAlias
				}
				entry.ParallelSlots[slotName] = se
				hasSlotPages = true
			}
		}

		if hasSlotPages {
			// For layout-only routes with parallel slots, we generate a route
			// with the first slot as the main component
			if route.LayoutFile != "" {
				entry.HasLayout = true
				entry.LayoutAlias = currentLayoutAlias
			}
			*entries = append(*entries, entry)
		}
	}

	// Recurse into children
	for _, child := range route.Children {
		collectRoutes(child, moduleName, imports, entries, currentLayoutAlias)
	}

	// Recurse into parallel slot children
	for _, slot := range route.ParallelSlots {
		for _, child := range slot.Children {
			collectRoutes(child, moduleName, imports, entries, currentLayoutAlias)
		}
	}
}

// dirToImportPath converts a directory path to a Go import path.
// It uses the directory path relative to the module root.
// For example: /app/about → myapp/app/about
func dirToImportPath(dirPath string, moduleName string) string {
	// Find the "app" segment and build from there
	// We assume the dirPath structure has a recognizable pattern
	// Use filepath.Base and parent parts to build the import path

	// Split the path into segments
	parts := splitPath(dirPath)

	// Find the index of the module-relative root (we take the last component matching common patterns)
	// Since we don't have the actual module root, we use a heuristic:
	// take the path from the "app" directory onward, relative to the module
	importSuffix := buildImportSuffix(parts)

	return moduleName + "/" + importSuffix
}

// buildImportSuffix builds the import path suffix from directory path parts.
// It takes the path starting from the "app" directory (or similar root).
func buildImportSuffix(parts []string) string {
	// Look for "app" as the root directory marker
	for i, part := range parts {
		if part == "app" {
			return strings.Join(parts[i:], "/")
		}
	}

	// Fallback: if no "app" found, use the last segment
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "root"
}

// splitPath splits a filepath into its components.
func splitPath(p string) []string {
	var parts []string
	for p != "" && p != "/" && p != "." {
		dir, file := filepath.Split(p)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		p = strings.TrimSuffix(dir, string(filepath.Separator))
	}
	return parts
}

// dirToAlias generates a Go-safe import alias from a directory path.
func dirToAlias(dirPath string) string {
	base := filepath.Base(dirPath)

	// Handle root directory
	if base == "." || base == "/" || base == "" {
		return "page"
	}

	return sanitizeAlias(base)
}

// sanitizeAlias converts a directory name to a valid Go identifier.
func sanitizeAlias(name string) string {
	// Remove special characters and convert to valid Go identifier
	var result strings.Builder
	for i, ch := range name {
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z':
			result.WriteRune(ch)
		case ch >= '0' && ch <= '9':
			if i == 0 {
				result.WriteString("p")
			}
			result.WriteRune(ch)
		case ch == '_':
			result.WriteRune(ch)
		default:
			// Skip special chars like [], (), @, .
			// But insert underscore if we have accumulated chars
			if result.Len() > 0 {
				result.WriteString("_")
			}
		}
	}

	s := result.String()
	s = strings.TrimRight(s, "_")

	if s == "" {
		return "page"
	}
	return s
}

// ensureUniqueAlias ensures the alias is unique among existing imports.
func ensureUniqueAlias(alias string, importPath string, imports map[string]routeImport) string {
	if alias == "" {
		alias = "page"
	}

	// Check if this alias is already used for a different import path
	candidate := alias
	counter := 2
	for {
		conflict := false
		for _, existing := range imports {
			if existing.Alias == candidate && existing.ImportPath != importPath {
				conflict = true
				break
			}
		}
		if !conflict {
			return candidate
		}
		candidate = fmt.Sprintf("%s%d", alias, counter)
		counter++
	}
}

// sortImports returns a sorted slice of route imports.
func sortImports(imports map[string]routeImport) []routeImport {
	result := make([]routeImport, 0, len(imports))
	for _, imp := range imports {
		result = append(result, imp)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ImportPath < result[j].ImportPath
	})
	return result
}
