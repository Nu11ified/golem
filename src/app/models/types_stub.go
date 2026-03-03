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
	ActivePageID string            `json:"activePageId"`
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
		Icon:      "\xf0\x9f\x93\x84",
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
