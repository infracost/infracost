package azure

import (
	"fmt"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getMSSQLElasticPoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_mssql_elasticpool",
		RFunc: newMSSQLElasticPool,
		ReferenceAttributes: []string{
			"server_name",
			"resource_group_name",
		},
	}
}

func newMSSQLElasticPool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := lookupRegion(d, []string{"server_name", "resource_group_name"})

	sku := d.Get("sku.0.name").String()
	capacity := d.Get("sku.0.capacity").Int()
	tier := strings.Join(camelcase.Split(d.Get("sku.0.tier").String()), " ")
	family := fmt.Sprintf("Compute %s", d.Get("sku.0.family").String())

	var maxSizeGB float64
	if !d.IsEmpty("max_size_gb") {
		maxSizeGB = d.Get("max_size_gb").Float()
	}
	if !d.IsEmpty("max_size_bytes") {
		maxSizeGB = d.Get("max_size_bytes").Float() / 1024.0 / 1024.0 / 1024.0
	}

	licenseType := d.GetStringOrDefault("license_type", "LicenseIncluded")

	r := &azure.MSSQLElasticPool{
		Address:       d.Address,
		Region:        region,
		SKU:           sku,
		Tier:          tier,
		Family:        family,
		LicenseType:   licenseType,
		MaxSizeGB:     &maxSizeGB,
		ZoneRedundant: d.Get("zone_redundant").Bool(),
	}

	s := strings.ToLower(r.SKU)
	if s == "basicpool" || s == "standardpool" || s == "premiumpool" {
		r.DTUCapacity = &capacity
	} else {
		r.Cores = &capacity
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
