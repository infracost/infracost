package google

import (
	"fmt"
	"regexp"

	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"

	"github.com/shopspring/decimal"
)

func GetContainerClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_container_cluster",
		RFunc: NewContainerCluster,
		Notes: []string{
			"Sustained use discounts are applied to monthly costs, but not to hourly costs.",
			"Costs associated with non-standard Linux images, such as Windows and RHEL are not supported.",
			"Custom machine types are not supported.",
		},
	}
}

func NewContainerCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	description := "Regional Kubernetes Clusters"
	region := d.Get("location").String()
	if isZone(region) {
		region = zoneToRegion(region)
		description = "Zonal Kubernetes Clusters"
	}

	subResources := make([]*schema.Resource, 0)

	if !d.Get("remove_default_node_pool").Bool() {
		zones := int64(zoneCount(d.RawValues, ""))

		countPerZone := int64(3)

		if d.Get("initial_node_count").Type != gjson.Null {
			countPerZone = d.Get("initial_node_count").Int()
		}

		if u != nil && u.Get("nodes").Exists() {
			countPerZone = u.Get("nodes").Int()
		}

		nodeCount := decimal.NewFromInt(zones * countPerZone)

		defaultPool := &schema.Resource{
			Name:           "default_pool",
			CostComponents: nodePoolCostComponents(region, d.Get("node_config.0")),
		}

		schema.MultiplyQuantities(defaultPool, nodeCount)

		subResources = append(subResources, defaultPool)
	}

	for i, values := range d.Get("node_pool").Array() {
		var countPerZoneOverride *int64
		k := fmt.Sprintf("node_pool[%d].nodes", i)
		if u != nil && u.Get(k).Exists() {
			c := u.Get(k).Int()
			countPerZoneOverride = &c
		}

		nodePool := newNodePool(fmt.Sprintf("node_pool[%d]", i), values, countPerZoneOverride, d)
		if nodePool != nil {
			subResources = append(subResources, nodePool)
		}
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           "Cluster management fee",
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr("global"),
				Service:       strPtr("Kubernetes Engine"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(description)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("0"),
				EndUsageAmount:   strPtr(""),
			},
		},
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func isZone(location string) bool {
	if matched, _ := regexp.MatchString(`^\w+-\w+-\w+$`, location); matched {
		return true
	}
	return false
}
