//go:build !js || !wasm

package dom

// Stub implementations for advanced Virtual DOM features
type VNode struct {
	Type      string
	Props     map[string]interface{}
	Children  []*VNode
	Key       string
	Component interface{}
	Hooks     *HookState
	IsDirty   bool
}

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

type VirtualDOM struct {
	Root       *VNode
	Components map[string]interface{}
	Scheduler  *Scheduler
}

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

// Stub functions
func NewVirtualDOM() *VirtualDOM {
	return &VirtualDOM{
		Components: make(map[string]interface{}),
		Scheduler: &Scheduler{
			UpdateQueue: make([]*VNode, 0),
			Priority:    NormalPriority,
		},
	}
}

func CreateVNode(nodeType string, props map[string]interface{}, children ...*VNode) *VNode {
	return &VNode{
		Type:     nodeType,
		Props:    props,
		Children: children,
		Hooks:    &HookState{},
		IsDirty:  true,
	}
}

func (vdom *VirtualDOM) Diff(oldTree, newTree *VNode) []Diff {
	return []Diff{}
}

func (vdom *VirtualDOM) Patch(diffs []Diff) {
	// No-op for non-WASM builds
}

func (vdom *VirtualDOM) Schedule(vnode *VNode, priority Priority) {
	// No-op for non-WASM builds
}
