package terraform

import (
	"sync"

	"github.com/infracost/infracost/pkg/schema"

	"github.com/infracost/infracost/internal/providers/terraform/aws"
)

type resourceRegistrySingleton map[string]schema.ResourceFunc

var (
	resourceRegistry resourceRegistrySingleton
	once             sync.Once
)

func getResourceRegistry() *resourceRegistrySingleton {
	once.Do(func() {
		resourceRegistry = make(resourceRegistrySingleton)
		// Merge all resource registries

		// AWS
		for k, v := range aws.ResourceRegistry {
			resourceRegistry[k] = v
		}

	})
	return &resourceRegistry
}
