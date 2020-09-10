package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

func NewDocDBClusterInstance(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	instanceType := d.Get("instance_class").String()

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Database Instance (%s, %s)", "on_demand", instanceType),
			Unit:           "hours",
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("Database Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "instanceType", Value: strPtr(instanceType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
		{
			Name: "Storage",
			Unit: "GB-months",
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("Database Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("StorageUsage")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
		{
			Name: "I/O",
			Unit: "requests",
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("StorageIOUsage")},
				},
			},
		},
		{
			Name: "Backup Storage",
			Unit: "GB-months",
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("Storage Snapshot"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("BackupUsage")},
				},
			},
		},
	}

	if strings.HasPrefix(instanceType, "db.t3.") {
		costComponents = append(costComponents, &schema.CostComponent{
			Name: "CPU Credits",
			Unit: "vCPU-hours",
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonDocDB"),
				ProductFamily: strPtr("CPU Credits"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("CPUCredits:db.t3")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
