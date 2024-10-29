package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func getSQLElasticPoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "azurerm_sql_elasticpool",
		CoreRFunc: newSQLElasticPool,
		ReferenceAttributes: []string{
			"server_name",
			"resource_group_name",
		},
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			return lookupRegion(d, []string{"server_name", "resource_group_name"})
		},
	}
}

func newSQLElasticPool(d *schema.ResourceData) schema.CoreResource {
	tier := d.Get("edition").String()
	sku := fmt.Sprintf("%sPool", strings.ToTitle(tier))
	dtu := d.Get("dtu").Int()

	region := d.Region
	r := &azure.MSSQLElasticPool{
		Address:       d.Address,
		Region:        region,
		SKU:           sku,
		Family:        "",
		Tier:          tier,
		DTUCapacity:   &dtu,
		LicenseType:   "LicenseIncluded",
		ZoneRedundant: d.Get("zone_redundant").Bool(),
	}

	if !d.IsEmpty("pool_size") {
		maxSizeGB := d.Get("pool_size").Float() / 1024.0
		r.MaxSizeGB = &maxSizeGB
	}

	return r
}
