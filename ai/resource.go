// File: ai/resource_manager.go

package ai

import (
	"fmt"

	"github.com/theapemachine/amsh/ai/tools"
)

type Resource interface {
	Use(string) (string, error)
}

type ResourceManager struct {
	resources map[string]Resource
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: map[string]Resource{
			"RESEARCH": tools.NewBrowser(),
		},
	}
}

func (rm *ResourceManager) UseResource(resourceType, query string) string {
	fmt.Printf("ResourceManager.UseResource called with type: %s, query: %s\n", resourceType, query)
	res, exists := rm.resources[resourceType]
	if !exists {
		return "Resource not found"
	}
	output, err := res.Use(query)
	if err != nil {
		return "Error: " + err.Error()
	}
	return output
}
