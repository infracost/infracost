package scaleway

import "github.com/infracost/infracost/internal/schema"

var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetInstanceIPRegistryItem(),
	GetInstanceServerRegistryItem(),
}

var FreeResources = []string{}

var UsageOnlyResources = []string{}
