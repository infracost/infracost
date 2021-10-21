package aws

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

var repUnderscore = regexp.MustCompile(`_`)

func GetDXConnectionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_dx_connection",
		RFunc: NewDXConnection,
	}
}

func NewDXConnection(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	virtualInterfaceType := "private"
	if u != nil && u.Get("dx_virtual_interface_type").Exists() {
		virtualInterfaceType = u.Get("dx_virtual_interface_type").String()
	}

	dxBandwidth := strings.Replace(d.Get("bandwidth").String(), "bps", "", 1)
	dxLocation := d.Get("location").String()

	connectionType := "dedicated"
	if u != nil && u.Get("dx_connection_type").Exists() {
		connectionType = u.Get("dx_connection_type").String()
	}

	components := []*schema.CostComponent{
		{
			Name:           "DX connection",
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AWSDirectConnect"),
				ProductFamily: strPtr("Direct Connect"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "capacity", Value: strPtr(dxBandwidth)},
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", dxLocation))},
					{Key: "connectionType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", connectionType))},
				},
			},
		},
	}

	if u != nil {
		hasPerRegionUsage := u.Get("monthly_outbound_from_region_to_dx_connection_location").Exists()

		if hasPerRegionUsage {
			regionUsage := u.Get("monthly_outbound_from_region_to_dx_connection_location").Map()

			// sort the region keys so that we get a consistent output in the cli
			keys := make([]string, 0, len(regionUsage))
			for key := range regionUsage {
				keys = append(keys, key)
			}

			sort.Strings(keys)

			for _, k := range keys {
				regionName := repUnderscore.ReplaceAllString(k, "-")
				fromLocation, ok := aws.RegionMapping[regionName]
				if !ok {
					log.Warnf("Skipping resource %s usage cost: Outbound data transfer. Could not find mapping for region %s", d.Address, k)
					continue
				}

				estimate := regionUsage[k]
				gbDataProcessed := decimalPtr(decimal.NewFromFloat(estimate.Float()))
				components = append(components, &schema.CostComponent{
					Name:            fmt.Sprintf("Outbound data transfer (from %s, to %s)", regionName, dxLocation),
					Unit:            "GB",
					UnitMultiplier:  decimal.NewFromInt(1),
					MonthlyQuantity: gbDataProcessed,
					ProductFilter: &schema.ProductFilter{
						VendorName:    strPtr("aws"),
						Service:       strPtr("AWSDirectConnect"),
						ProductFamily: strPtr("Data Transfer"),
						AttributeFilters: []*schema.AttributeFilter{
							{Key: "fromLocation", Value: strPtr(fromLocation)},
							{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s-DataXfer-Out/", dxLocation))},
							{Key: "virtualInterfaceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", virtualInterfaceType))},
						},
					},
				})
			}
		}

		// if we don't have any dx_connection usage data using the new
		// "monthly_outbound_from_region_to_dx_connection_location" yaml conf
		// we should fall back to the old "monthly_outbound_region_to_dx_location_gb" configuration
		if !hasPerRegionUsage && u.Get("monthly_outbound_region_to_dx_location_gb").Exists() {
			fromLocation, ok := aws.RegionMapping[region]
			if ok {
				gbDataProcessed := decimalPtr(decimal.NewFromFloat(u.Get("monthly_outbound_region_to_dx_location_gb").Float()))
				components = append(components, &schema.CostComponent{
					Name:            fmt.Sprintf("Outbound data transfer (from %s, to %s)", region, dxLocation),
					Unit:            "GB",
					UnitMultiplier:  decimal.NewFromInt(1),
					MonthlyQuantity: gbDataProcessed,
					ProductFilter: &schema.ProductFilter{
						VendorName:    strPtr("aws"),
						Service:       strPtr("AWSDirectConnect"),
						ProductFamily: strPtr("Data Transfer"),
						AttributeFilters: []*schema.AttributeFilter{
							{Key: "fromLocation", Value: strPtr(fromLocation)},
							{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s-DataXfer-Out/", dxLocation))},
							{Key: "virtualInterfaceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", virtualInterfaceType))},
						},
					},
				})
			}
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: components,
	}
}
