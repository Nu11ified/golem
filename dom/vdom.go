//go:build js && wasm

// Advanced Virtual DOM with diffing algorithms
package dom

import (
	"reflect"
	"syscall/js"
)

// VNode represents a virtual DOM node with diffing capabilities
type VNode struct {
	Type      string
	Props     map[string]interface{}
	Children  []*VNode
	Key       string      // For optimal list diffing
	Component interface{} // Component reference
	Hooks     *HookState  // React-like hooks
	JSElement js.Value
	IsDirty   bool
}

// HookState manages component state and effects
type HookState struct {
	StateIndex  int
	EffectIndex int
	States      []interface{}
	Effects     []Effect
	Deps        [][]interface{}
}

type Effect struct {
	Fn      func()
	Cleanup func()
	Deps    []interface{}
}

// Diff represents a change in the virtual DOM
type Diff struct {
	Type    DiffType
	OldNode *VNode
	NewNode *VNode
	Index   int
	Props   map[string]interface{}
}

type DiffType int

const (
	DiffCreate DiffType = iota
	DiffUpdate
	DiffRemove
	DiffReplace
	DiffReorder
)

// VirtualDOM manages the virtual DOM tree and diffing
type VirtualDOM struct {
	Root       *VNode
	Components map[string]interface{}
	Scheduler  *Scheduler
}

// Scheduler manages rendering updates efficiently
type Scheduler struct {
	UpdateQueue []*VNode
	IsScheduled bool
	Priority    Priority
}

type Priority int

const (
	ImmediatePriority Priority = iota
	UserBlockingPriority
	NormalPriority
	LowPriority
	IdlePriority
)

// NewVirtualDOM creates a new virtual DOM instance
func NewVirtualDOM() *VirtualDOM {
	return &VirtualDOM{
		Components: make(map[string]interface{}),
		Scheduler: &Scheduler{
			UpdateQueue: make([]*VNode, 0),
			Priority:    NormalPriority,
		},
	}
}

// CreateVNode creates a new virtual node
func CreateVNode(nodeType string, props map[string]interface{}, children ...*VNode) *VNode {
	vnode := &VNode{
		Type:     nodeType,
		Props:    props,
		Children: children,
		Hooks:    &HookState{States: make([]interface{}, 0), Effects: make([]Effect, 0)},
		IsDirty:  true,
	}

	// Extract key if provided
	if key, ok := props["key"]; ok {
		vnode.Key = key.(string)
	}

	return vnode
}

// Diff compares two virtual DOM trees and returns differences
func (vdom *VirtualDOM) Diff(oldTree, newTree *VNode) []Diff {
	diffs := make([]Diff, 0)
	vdom.diffRecursive(oldTree, newTree, &diffs, 0)
	return diffs
}

func (vdom *VirtualDOM) diffRecursive(oldNode, newNode *VNode, diffs *[]Diff, index int) {
	// Node removed
	if oldNode != nil && newNode == nil {
		*diffs = append(*diffs, Diff{
			Type:    DiffRemove,
			OldNode: oldNode,
			Index:   index,
		})
		return
	}

	// Node added
	if oldNode == nil && newNode != nil {
		*diffs = append(*diffs, Diff{
			Type:    DiffCreate,
			NewNode: newNode,
			Index:   index,
		})
		return
	}

	// Different node types - replace
	if oldNode.Type != newNode.Type {
		*diffs = append(*diffs, Diff{
			Type:    DiffReplace,
			OldNode: oldNode,
			NewNode: newNode,
			Index:   index,
		})
		return
	}

	// Same type - check props
	propDiffs := vdom.diffProps(oldNode.Props, newNode.Props)
	if len(propDiffs) > 0 {
		*diffs = append(*diffs, Diff{
			Type:    DiffUpdate,
			OldNode: oldNode,
			NewNode: newNode,
			Props:   propDiffs,
			Index:   index,
		})
	}

	// Diff children with key-based optimization
	vdom.diffChildren(oldNode.Children, newNode.Children, diffs, index)
}

// diffProps compares properties between nodes
func (vdom *VirtualDOM) diffProps(oldProps, newProps map[string]interface{}) map[string]interface{} {
	changes := make(map[string]interface{})

	// Check for changed/added props
	for key, newValue := range newProps {
		if oldValue, exists := oldProps[key]; !exists || !reflect.DeepEqual(oldValue, newValue) {
			changes[key] = newValue
		}
	}

	// Check for removed props
	for key := range oldProps {
		if _, exists := newProps[key]; !exists {
			changes[key] = nil // Mark as removed
		}
	}

	return changes
}

// diffChildren uses key-based diffing for optimal performance
func (vdom *VirtualDOM) diffChildren(oldChildren, newChildren []*VNode, diffs *[]Diff, parentIndex int) {
	// Simple case: no keys, diff by index
	if !vdom.hasKeys(oldChildren) && !vdom.hasKeys(newChildren) {
		maxLen := len(oldChildren)
		if len(newChildren) > maxLen {
			maxLen = len(newChildren)
		}

		for i := 0; i < maxLen; i++ {
			var oldChild, newChild *VNode
			if i < len(oldChildren) {
				oldChild = oldChildren[i]
			}
			if i < len(newChildren) {
				newChild = newChildren[i]
			}
			vdom.diffRecursive(oldChild, newChild, diffs, i)
		}
		return
	}

	// Key-based diffing for reordering optimization
	vdom.diffChildrenWithKeys(oldChildren, newChildren, diffs, parentIndex)
}

// hasKeys checks if any child has a key
func (vdom *VirtualDOM) hasKeys(children []*VNode) bool {
	for _, child := range children {
		if child != nil && child.Key != "" {
			return true
		}
	}
	return false
}

// diffChildrenWithKeys implements efficient key-based diffing
func (vdom *VirtualDOM) diffChildrenWithKeys(oldChildren, newChildren []*VNode, diffs *[]Diff, parentIndex int) {
	oldKeyMap := make(map[string]int)
	newKeyMap := make(map[string]int)

	// Build key maps
	for i, child := range oldChildren {
		if child != nil && child.Key != "" {
			oldKeyMap[child.Key] = i
		}
	}

	for i, child := range newChildren {
		if child != nil && child.Key != "" {
			newKeyMap[child.Key] = i
		}
	}

	// Track moves and changes
	moves := make([]int, len(newChildren))
	for i := range moves {
		moves[i] = -1
	}

	// Find matching keys
	for newIndex, newChild := range newChildren {
		if newChild != nil && newChild.Key != "" {
			if oldIndex, exists := oldKeyMap[newChild.Key]; exists {
				moves[newIndex] = oldIndex
				vdom.diffRecursive(oldChildren[oldIndex], newChild, diffs, newIndex)
			} else {
				// New node
				*diffs = append(*diffs, Diff{
					Type:    DiffCreate,
					NewNode: newChild,
					Index:   newIndex,
				})
			}
		}
	}

	// Handle reordering
	if vdom.needsReorder(moves) {
		*diffs = append(*diffs, Diff{
			Type:  DiffReorder,
			Index: parentIndex,
		})
	}

	// Handle removed nodes
	for oldIndex, oldChild := range oldChildren {
		if oldChild != nil && oldChild.Key != "" {
			if _, exists := newKeyMap[oldChild.Key]; !exists {
				*diffs = append(*diffs, Diff{
					Type:    DiffRemove,
					OldNode: oldChild,
					Index:   oldIndex,
				})
			}
		}
	}
}

// needsReorder checks if the moves array indicates reordering is needed
func (vdom *VirtualDOM) needsReorder(moves []int) bool {
	lastIndex := -1
	for _, moveIndex := range moves {
		if moveIndex != -1 {
			if moveIndex < lastIndex {
				return true
			}
			lastIndex = moveIndex
		}
	}
	return false
}

// Patch applies diffs to the actual DOM
func (vdom *VirtualDOM) Patch(diffs []Diff) {
	for _, diff := range diffs {
		switch diff.Type {
		case DiffCreate:
			vdom.createElement(diff.NewNode)
		case DiffUpdate:
			vdom.updateElement(diff.NewNode, diff.Props)
		case DiffRemove:
			vdom.removeElement(diff.OldNode)
		case DiffReplace:
			vdom.replaceElement(diff.OldNode, diff.NewNode)
		case DiffReorder:
			vdom.reorderChildren(diff.OldNode, diff.NewNode)
		}
	}
}

// createElement creates a new DOM element
func (vdom *VirtualDOM) createElement(vnode *VNode) {
	if vnode.JSElement.IsUndefined() {
		doc := js.Global().Get("document")
		vnode.JSElement = doc.Call("createElement", vnode.Type)

		// Set properties
		for name, value := range vnode.Props {
			vdom.setProperty(vnode.JSElement, name, value)
		}

		// Create children
		for _, child := range vnode.Children {
			vdom.createElement(child)
			vnode.JSElement.Call("appendChild", child.JSElement)
		}
	}
}

// updateElement updates an existing DOM element
func (vdom *VirtualDOM) updateElement(vnode *VNode, propChanges map[string]interface{}) {
	if !vnode.JSElement.IsUndefined() {
		for name, value := range propChanges {
			vdom.setProperty(vnode.JSElement, name, value)
		}
	}
}

// removeElement removes a DOM element
func (vdom *VirtualDOM) removeElement(vnode *VNode) {
	if !vnode.JSElement.IsUndefined() {
		parent := vnode.JSElement.Get("parentNode")
		if !parent.IsNull() {
			parent.Call("removeChild", vnode.JSElement)
		}
	}
}

// replaceElement replaces one DOM element with another
func (vdom *VirtualDOM) replaceElement(oldNode, newNode *VNode) {
	vdom.createElement(newNode)
	if !oldNode.JSElement.IsUndefined() {
		parent := oldNode.JSElement.Get("parentNode")
		if !parent.IsNull() {
			parent.Call("replaceChild", newNode.JSElement, oldNode.JSElement)
		}
	}
}

// reorderChildren reorders child elements
func (vdom *VirtualDOM) reorderChildren(oldNode, newNode *VNode) {
	// Implementation for reordering - complex DOM manipulation
	// This would involve moving actual DOM nodes to match new order
}

// setProperty sets a property on a DOM element
func (vdom *VirtualDOM) setProperty(element js.Value, name string, value interface{}) {
	switch name {
	case "className":
		element.Set("className", value)
	case "textContent":
		element.Set("textContent", value)
	case "innerHTML":
		element.Set("innerHTML", value)
	case "value":
		element.Set("value", value)
	case "checked":
		element.Set("checked", value)
	case "disabled":
		element.Set("disabled", value)
	default:
		if value == nil {
			element.Call("removeAttribute", name)
		} else {
			element.Call("setAttribute", name, value)
		}
	}
}

// Schedule queues a component for re-rendering
func (vdom *VirtualDOM) Schedule(vnode *VNode, priority Priority) {
	vdom.Scheduler.UpdateQueue = append(vdom.Scheduler.UpdateQueue, vnode)
	vdom.Scheduler.Priority = priority

	if !vdom.Scheduler.IsScheduled {
		vdom.Scheduler.IsScheduled = true
		vdom.flushWork()
	}
}

// flushWork processes the update queue
func (vdom *VirtualDOM) flushWork() {
	// Use requestIdleCallback for low priority updates
	if vdom.Scheduler.Priority == LowPriority || vdom.Scheduler.Priority == IdlePriority {
		callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			vdom.processUpdates()
			return nil
		})
		js.Global().Call("requestIdleCallback", callback)
	} else {
		// Use requestAnimationFrame for higher priority updates
		callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			vdom.processUpdates()
			return nil
		})
		js.Global().Call("requestAnimationFrame", callback)
	}
}

// processUpdates processes all queued updates
func (vdom *VirtualDOM) processUpdates() {
	for len(vdom.Scheduler.UpdateQueue) > 0 {
		vnode := vdom.Scheduler.UpdateQueue[0]
		vdom.Scheduler.UpdateQueue = vdom.Scheduler.UpdateQueue[1:]

		if vnode.IsDirty {
			vdom.renderComponent(vnode)
			vnode.IsDirty = false
		}
	}
	vdom.Scheduler.IsScheduled = false
}

// renderComponent renders a single component
func (vdom *VirtualDOM) renderComponent(vnode *VNode) {
	// Component rendering logic would go here
	// This would call the component function and diff the result
}

// Concurrent features for future enhancement
type Fiber struct {
	VNode    *VNode
	Parent   *Fiber
	Child    *Fiber
	Sibling  *Fiber
	WorkType WorkType
	Priority Priority
}

type WorkType int

const (
	WorkUpdate WorkType = iota
	WorkInsert
	WorkDelete
)
