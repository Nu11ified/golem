package server

import (
	"log"

	"github.com/Nu11ified/golem/functions"
)

var defaultStorage *Storage

func init() {
	defaultStorage = NewStorage("data")

	mustRegister := func(namespace, name string, fn interface{}) {
		if err := functions.Register(namespace, name, fn); err != nil {
			log.Fatalf("failed to register %s.%s: %v", namespace, name, err)
		}
	}

	mustRegister("notion", "GetWorkspace", func() (interface{}, error) {
		return defaultStorage.LoadWorkspace()
	})
	mustRegister("notion", "SaveWorkspace", func(data interface{}) error {
		return defaultStorage.SaveWorkspace(data)
	})
	mustRegister("notion", "GetPageBlocks", func(pageID string) ([]interface{}, error) {
		return defaultStorage.LoadPageBlocks(pageID)
	})
	mustRegister("notion", "SavePageBlocks", func(pageID string, blocks interface{}) error {
		return defaultStorage.SavePageBlocks(pageID, blocks)
	})
	mustRegister("notion", "DeletePageData", func(pageID string) error {
		return defaultStorage.DeletePageData(pageID)
	})
}
