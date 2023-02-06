package azurerm

import (
	"github.com/infracost/infracost/internal/schema"
)

// TODO: see providers/terraform/azure/registry.go
// Use ASP/Function app as testing resources
var Registry []*schema.RegistryItem = []*schema.RegistryItem{
	getAppFunctionRegistryItem(),
	getAppServicePlanRegistryItem(),
}

type ResourceRegistryMap map[string]*schema.RegistryItem

// TODO: see providers/terraform/azure/registry.go
var FreeResources = []string{}

var UsageOnlyResources = []string{}

func GetRegistryMap() *ResourceRegistryMap {
}
