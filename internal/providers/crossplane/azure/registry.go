package azure

import "github.com/infracost/infracost/internal/schema"

// ResourceRegistry grouped alphabetically
var ResourceRegistry []*schema.RegistryItem = []*schema.RegistryItem{
	GetAzureRMKubernetesClusterRegistryItem(),
}

// FreeResources grouped alphabetically
var FreeResources = []string{}

// UsageOnlyResources grouped alphabetically
var UsageOnlyResources = []string{}
