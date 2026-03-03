package server

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Storage handles JSON file persistence for the Notion clone
type Storage struct {
	dataDir string
}

// NewStorage creates a new storage instance
func NewStorage(dataDir string) *Storage {
	return &Storage{dataDir: dataDir}
}

// SaveWorkspace writes workspace data to workspace.json
func (s *Storage) SaveWorkspace(data interface{}) error {
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(s.dataDir, "workspace.json"))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(data)
}

// LoadWorkspace reads workspace data from workspace.json
func (s *Storage) LoadWorkspace() (interface{}, error) {
	path := filepath.Join(s.dataDir, "workspace.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SavePageBlocks writes block data to blocks/<pageId>.json
func (s *Storage) SavePageBlocks(pageID string, blocks interface{}) error {
	dir := filepath.Join(s.dataDir, "blocks")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, pageID+".json"))
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(blocks)
}

// LoadPageBlocks reads block data from blocks/<pageId>.json
func (s *Storage) LoadPageBlocks(pageID string) ([]interface{}, error) {
	path := filepath.Join(s.dataDir, "blocks", pageID+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var result []interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeletePageData removes the blocks file for a page
func (s *Storage) DeletePageData(pageID string) error {
	path := filepath.Join(s.dataDir, "blocks", pageID+".json")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
