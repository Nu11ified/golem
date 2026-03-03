package codegen

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ScannedRoute represents a discovered route from the filesystem
type ScannedRoute struct {
	Path               string                   // URL path pattern (e.g., "/blog/:slug")
	Segment            string                   // directory name
	DirPath            string                   // absolute filesystem path to directory
	PageFile           string                   // path to page.go (if exists)
	LayoutFile         string                   // path to layout.go (if exists)
	ErrorFile          string                   // path to error.go (if exists)
	LoadingFile        string                   // path to loading.go (if exists)
	NotFoundFile       string                   // path to notfound.go (if exists)
	TemplateFile       string                   // path to template.go (if exists)
	Children           []*ScannedRoute
	ParamName          string                   // for [slug] -> "slug"
	IsCatchAll         bool                     // for [...path]
	IsOptionalCatchAll bool                     // for [[...path]]
	IsGroup            bool                     // for (marketing) -- not in URL
	GroupName          string                   // "marketing" for (marketing)
	ParallelSlots      map[string]*ScannedRoute // for @sidebar -> "sidebar"
	InterceptPattern   string                   // "(.)", "(..)", "(...)"
	InterceptDepth     int                      // 0=same level, 1=parent, -1=root
}

// Regex patterns for directory naming conventions
var (
	// [param] - dynamic segment
	dynamicSegmentRe = regexp.MustCompile(`^\[([a-zA-Z_][a-zA-Z0-9_]*)\]$`)
	// [...catchAll] - catch-all segment
	catchAllRe = regexp.MustCompile(`^\[\.\.\.([a-zA-Z_][a-zA-Z0-9_]*)\]$`)
	// [[...optional]] - optional catch-all segment
	optionalCatchAllRe = regexp.MustCompile(`^\[\[\.\.\.([a-zA-Z_][a-zA-Z0-9_]*)\]\]$`)
	// (group) - route group (not a route segment, not in URL)
	routeGroupRe = regexp.MustCompile(`^\(([a-zA-Z_][a-zA-Z0-9_]*)\)$`)
	// @slot - parallel route slot
	parallelSlotRe = regexp.MustCompile(`^@([a-zA-Z_][a-zA-Z0-9_]*)$`)
	// (.)route, (..)route, (...)route - intercepting routes
	interceptingRouteRe = regexp.MustCompile(`^(\(\.{1,3}\))(.+)$`)
)

// Special file names that are recognized in each directory
var specialFiles = []string{
	"page.go",
	"layout.go",
	"error.go",
	"loading.go",
	"notfound.go",
	"template.go",
}

// ScanRoutes scans the app directory and returns a route tree
func ScanRoutes(appDir string) (*ScannedRoute, error) {
	absDir, err := filepath.Abs(appDir)
	if err != nil {
		return nil, err
	}

	root := &ScannedRoute{
		Path:    "/",
		Segment: "",
		DirPath: absDir,
	}

	// Detect special files in the root directory
	detectSpecialFiles(root, absDir)

	// Recursively scan children
	if err := scanChildren(root, absDir, "/"); err != nil {
		return nil, err
	}

	return root, nil
}

// scanChildren reads directory entries and recurses into subdirectories
func scanChildren(parent *ScannedRoute, dirPath string, parentURLPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Sort entries for deterministic output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden directories
		if strings.HasPrefix(name, ".") {
			continue
		}

		childDirPath := filepath.Join(dirPath, name)

		// Check if this is a parallel slot (@sidebar)
		if matches := parallelSlotRe.FindStringSubmatch(name); matches != nil {
			slotName := matches[1]
			slotRoute := &ScannedRoute{
				Segment: name,
				DirPath: childDirPath,
				Path:    parentURLPath, // parallel slots share the parent path
			}
			detectSpecialFiles(slotRoute, childDirPath)

			if parent.ParallelSlots == nil {
				parent.ParallelSlots = make(map[string]*ScannedRoute)
			}
			parent.ParallelSlots[slotName] = slotRoute

			// Recurse into the slot's children
			if err := scanChildren(slotRoute, childDirPath, parentURLPath); err != nil {
				return err
			}
			continue
		}

		child := &ScannedRoute{
			Segment: name,
			DirPath: childDirPath,
		}

		// Parse directory name to determine route type
		parseDirName(child, name)

		// Build the URL path for this segment
		child.Path = buildURLPath(parentURLPath, child)

		// Detect special files
		detectSpecialFiles(child, childDirPath)

		// Add as child of parent
		parent.Children = append(parent.Children, child)

		// Recurse into this directory's children
		childURLPath := child.Path
		if child.IsGroup {
			// Groups don't contribute to the URL, pass parent's URL path through
			childURLPath = parentURLPath
		}
		if err := scanChildren(child, childDirPath, childURLPath); err != nil {
			return err
		}
	}

	return nil
}

// parseDirName parses the directory name and sets route type fields
func parseDirName(route *ScannedRoute, name string) {
	// Check intercepting route first (must come before route group check
	// since intercepting patterns also use parentheses)
	if matches := interceptingRouteRe.FindStringSubmatch(name); matches != nil {
		route.InterceptPattern = matches[1]
		routeName := matches[2]

		// Determine intercept depth
		switch matches[1] {
		case "(.)":
			route.InterceptDepth = 0
		case "(..)":
			route.InterceptDepth = 1
		case "(...)":
			route.InterceptDepth = -1
		}

		// The route name part may itself be a dynamic segment
		parseDirName(route, routeName)
		// If it wasn't a special pattern, set the segment to the route name
		if route.Segment == name {
			route.Segment = routeName
		}
		return
	}

	// Optional catch-all: [[...param]]
	if matches := optionalCatchAllRe.FindStringSubmatch(name); matches != nil {
		route.ParamName = matches[1]
		route.IsOptionalCatchAll = true
		route.IsCatchAll = true
		return
	}

	// Catch-all: [...param]
	if matches := catchAllRe.FindStringSubmatch(name); matches != nil {
		route.ParamName = matches[1]
		route.IsCatchAll = true
		return
	}

	// Dynamic segment: [param]
	if matches := dynamicSegmentRe.FindStringSubmatch(name); matches != nil {
		route.ParamName = matches[1]
		return
	}

	// Route group: (group)
	if matches := routeGroupRe.FindStringSubmatch(name); matches != nil {
		route.IsGroup = true
		route.GroupName = matches[1]
		return
	}
}

// buildURLPath constructs the URL path for a route segment
func buildURLPath(parentPath string, route *ScannedRoute) string {
	// Route groups don't add to the URL path
	if route.IsGroup {
		return parentPath
	}

	// Determine the URL segment
	var segment string
	switch {
	case route.IsOptionalCatchAll:
		segment = "*" + route.ParamName
	case route.IsCatchAll:
		segment = "*" + route.ParamName
	case route.ParamName != "":
		segment = ":" + route.ParamName
	case route.InterceptPattern != "":
		// For intercepting routes, use the segment name (already parsed)
		segment = route.Segment
	default:
		segment = route.Segment
	}

	// Build the full path
	if parentPath == "/" {
		return "/" + segment
	}
	return parentPath + "/" + segment
}

// detectSpecialFiles checks for the presence of special files in a directory
func detectSpecialFiles(route *ScannedRoute, dirPath string) {
	for _, filename := range specialFiles {
		fullPath := filepath.Join(dirPath, filename)
		if _, err := os.Stat(fullPath); err == nil {
			switch filename {
			case "page.go":
				route.PageFile = fullPath
			case "layout.go":
				route.LayoutFile = fullPath
			case "error.go":
				route.ErrorFile = fullPath
			case "loading.go":
				route.LoadingFile = fullPath
			case "notfound.go":
				route.NotFoundFile = fullPath
			case "template.go":
				route.TemplateFile = fullPath
			}
		}
	}
}
