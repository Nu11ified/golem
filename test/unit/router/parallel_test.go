//go:build !js || !wasm

package router_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Nu11ified/golem/dom"
	"github.com/Nu11ified/golem/router"
)

// TestRouter_RenderParallelSlots_Basic verifies that a route with two
// parallel slots (sidebar and content) renders both independently and
// returns them in the resulting map.
func TestRouter_RenderParallelSlots_Basic(t *testing.T) {
	r := router.NewRouter()

	route := &router.Route{
		Path: "/dashboard",
		ParallelSlots: map[string]*router.Route{
			"sidebar": {
				Path: "/dashboard",
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Class("sidebar"), dom.P("sidebar content"))
				},
			},
			"content": {
				Path: "/dashboard",
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Class("content"), dom.P("main content"))
				},
			},
		},
	}

	r.AddRoute(route)

	slots := r.RenderParallelSlots(route, nil)

	// Should have exactly 2 slots (no "children" since route has no Component)
	if len(slots) != 2 {
		t.Errorf("expected 2 slots, got %d", len(slots))
	}

	// Verify sidebar slot rendered
	sidebar, ok := slots["sidebar"]
	if !ok {
		t.Fatal("expected sidebar slot in result")
	}
	if sidebar == nil {
		t.Fatal("expected sidebar slot element, got nil")
	}
	sidebarHTML := dom.RenderToHTML(sidebar)
	if !strings.Contains(sidebarHTML, "sidebar content") {
		t.Errorf("expected sidebar content in HTML, got %s", sidebarHTML)
	}

	// Verify content slot rendered
	content, ok := slots["content"]
	if !ok {
		t.Fatal("expected content slot in result")
	}
	if content == nil {
		t.Fatal("expected content slot element, got nil")
	}
	contentHTML := dom.RenderToHTML(content)
	if !strings.Contains(contentHTML, "main content") {
		t.Errorf("expected main content in HTML, got %s", contentHTML)
	}
}

// TestRouter_RenderParallelSlots_SlotError verifies that when one parallel
// slot panics, the other slot still renders successfully and the panicking
// slot's error boundary catches the error.
func TestRouter_RenderParallelSlots_SlotError(t *testing.T) {
	r := router.NewRouter()

	route := &router.Route{
		Path: "/dashboard",
		ParallelSlots: map[string]*router.Route{
			"sidebar": {
				Path: "/dashboard",
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Class("sidebar"), dom.P("sidebar ok"))
				},
			},
			"content": {
				Path: "/dashboard",
				Component: func(params map[string]string) *dom.Element {
					panic(errors.New("content slot exploded"))
				},
				ErrorHandler: func(err error) *dom.Element {
					return dom.Div(dom.Class("error"), dom.P(err.Error()))
				},
			},
		},
	}

	r.AddRoute(route)

	slots := r.RenderParallelSlots(route, nil)

	// Sidebar should render normally
	sidebar := slots["sidebar"]
	if sidebar == nil {
		t.Fatal("expected sidebar slot element, got nil")
	}
	sidebarHTML := dom.RenderToHTML(sidebar)
	if !strings.Contains(sidebarHTML, "sidebar ok") {
		t.Errorf("expected sidebar content in HTML, got %s", sidebarHTML)
	}

	// Content should have error boundary output
	content := slots["content"]
	if content == nil {
		t.Fatal("expected content error element, got nil")
	}
	contentHTML := dom.RenderToHTML(content)
	if !strings.Contains(contentHTML, "content slot exploded") {
		t.Errorf("expected error message in content HTML, got %s", contentHTML)
	}
	if !strings.Contains(contentHTML, "error") {
		t.Errorf("expected error class in content HTML, got %s", contentHTML)
	}
}

// TestRouter_RenderParallelSlots_EmptySlot verifies that a slot with no
// component results in nil in the returned map.
func TestRouter_RenderParallelSlots_EmptySlot(t *testing.T) {
	r := router.NewRouter()

	route := &router.Route{
		Path: "/dashboard",
		ParallelSlots: map[string]*router.Route{
			"sidebar": {
				Path: "/dashboard",
				// No Component — this slot is empty
			},
			"content": {
				Path: "/dashboard",
				Component: func(params map[string]string) *dom.Element {
					return dom.P("content here")
				},
			},
		},
	}

	r.AddRoute(route)

	slots := r.RenderParallelSlots(route, nil)

	// sidebar should be nil
	sidebar, ok := slots["sidebar"]
	if !ok {
		t.Fatal("expected sidebar key in result")
	}
	if sidebar != nil {
		t.Errorf("expected nil for empty slot, got element of type %s", sidebar.Type)
	}

	// content should render fine
	content := slots["content"]
	if content == nil {
		t.Fatal("expected content element, got nil")
	}
	contentHTML := dom.RenderToHTML(content)
	if !strings.Contains(contentHTML, "content here") {
		t.Errorf("expected content text in HTML, got %s", contentHTML)
	}
}

// TestRouter_RenderParallelSlots_DefaultChildren verifies that when a route
// has both a Component and parallel slots, the Component is rendered as the
// "children" slot in the returned map.
func TestRouter_RenderParallelSlots_DefaultChildren(t *testing.T) {
	r := router.NewRouter()

	route := &router.Route{
		Path: "/dashboard",
		Component: func(params map[string]string) *dom.Element {
			return dom.Div(dom.Class("main-page"), dom.P("default children"))
		},
		ParallelSlots: map[string]*router.Route{
			"sidebar": {
				Path: "/dashboard",
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Class("sidebar"), dom.P("sidebar content"))
				},
			},
		},
	}

	r.AddRoute(route)

	slots := r.RenderParallelSlots(route, nil)

	// Should have "children" + "sidebar" = 2 slots
	if len(slots) != 2 {
		t.Errorf("expected 2 slots, got %d", len(slots))
	}

	// Verify "children" slot (from main Component)
	children, ok := slots["children"]
	if !ok {
		t.Fatal("expected 'children' slot in result")
	}
	if children == nil {
		t.Fatal("expected children element, got nil")
	}
	childrenHTML := dom.RenderToHTML(children)
	if !strings.Contains(childrenHTML, "default children") {
		t.Errorf("expected 'default children' in HTML, got %s", childrenHTML)
	}

	// Verify sidebar still renders
	sidebar := slots["sidebar"]
	if sidebar == nil {
		t.Fatal("expected sidebar element, got nil")
	}
	sidebarHTML := dom.RenderToHTML(sidebar)
	if !strings.Contains(sidebarHTML, "sidebar content") {
		t.Errorf("expected sidebar content in HTML, got %s", sidebarHTML)
	}
}

// TestRouter_ParallelSlots_IndependentRendering verifies that parallel slots
// render independently — each slot receives its own params and they do not
// share mutable state.
func TestRouter_ParallelSlots_IndependentRendering(t *testing.T) {
	r := router.NewRouter()

	// Each slot's component captures a different piece of the params to
	// prove they receive the same params map but execute independently.
	route := &router.Route{
		Path: "/user/:id",
		ParallelSlots: map[string]*router.Route{
			"profile": {
				Path: "/user/:id",
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Class("profile"), dom.P("profile-"+params["id"]))
				},
			},
			"activity": {
				Path: "/user/:id",
				Component: func(params map[string]string) *dom.Element {
					return dom.Div(dom.Class("activity"), dom.P("activity-"+params["id"]))
				},
			},
		},
	}

	r.AddRoute(route)

	params := map[string]string{"id": "42"}
	slots := r.RenderParallelSlots(route, params)

	// Verify profile slot used params correctly
	profile := slots["profile"]
	if profile == nil {
		t.Fatal("expected profile slot element, got nil")
	}
	profileHTML := dom.RenderToHTML(profile)
	if !strings.Contains(profileHTML, "profile-42") {
		t.Errorf("expected profile-42 in HTML, got %s", profileHTML)
	}

	// Verify activity slot used params correctly and independently
	activity := slots["activity"]
	if activity == nil {
		t.Fatal("expected activity slot element, got nil")
	}
	activityHTML := dom.RenderToHTML(activity)
	if !strings.Contains(activityHTML, "activity-42") {
		t.Errorf("expected activity-42 in HTML, got %s", activityHTML)
	}

	// Verify the two slots produced different content (independent rendering)
	if profileHTML == activityHTML {
		t.Error("expected profile and activity slots to produce different HTML")
	}
}
