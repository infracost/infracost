package ibm

import (
	"fmt"
	"math"
	"strconv"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetElasticSearchCostComponents(r *Database) []*schema.CostComponent {

	if r.Flavor != "" && r.Flavor != "multitenant" {
		return []*schema.CostComponent{
			ElasticSearchHostFlavorComponent(r),
			ElasticSearchDiskCostComponent(r),
		}
	} else {
		return []*schema.CostComponent{
			ElasticSearchRAMCostComponent(r),
			ElasticSearchDiskCostComponent(r),
			ElasticSearchVirtualProcessorCoreCostComponent(r),
		}
	}
}

func ElasticSearchVirtualProcessorCoreCostComponent(r *Database) *schema.CostComponent {

	var q float64 = float64(r.CPU)
	if r.Flavor == "multitenant" && q == 0 {
		// Calculate CPU as 1:8 ratio with RAM, with a max of 2 CPU https://cloud.ibm.com/docs/databases-for-elasticsearch?topic=databases-for-elasticsearch-resources-scaling&interface=terraform
		q = math.Min(float64(r.Memory/1024)/8, 2)
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Virtual Processor Cores (%s members)", strconv.FormatInt(r.Members, 10)),
		Unit:            "CPU",
		UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(float64(q) * float64(r.Members))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr("databases-for-elasticsearch"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("VIRTUAL_PROCESSOR_CORES"),
		},
	}
}

func ElasticSearchRAMCostComponent(r *Database) *schema.CostComponent {

	unit := "GIGABYTE_MONTHS_RAM"
	if r.Plan == "enterprise" {
		unit = "GIGABYTE_MONTHS_RAM_NEW"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("RAM (%s members)", strconv.FormatInt(r.Members, 10)),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),                                                         // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(float64(r.Memory*r.Members) / float64(1024))), // Convert to GB
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr("databases-for-elasticsearch"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(unit),
		},
	}
}

func ElasticSearchDiskCostComponent(r *Database) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Disk (%s members)", strconv.FormatInt(r.Members, 10)),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),                                                       // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(float64(r.Disk*r.Members) / float64(1024))), // Convert to GB
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr("databases-for-elasticsearch"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_MONTHS_DISK"),
		},
	}
}

func ElasticSearchHostFlavorComponent(r *Database) *schema.CostComponent {

	CostUnit := map[string]string{
		"b3c.4x16.encrypted":   "HOST_FOUR_SIXTEEN",
		"b3c.8x32.encrypted":   "HOST_EIGHT_THIRTYTWO",
		"m3c.8x64.encrypted":   "HOST_EIGHT_SIXTYFOUR",
		"b3c.16x64.encrypted":  "HOST_SIXTEEN_SIXTYFOUR",
		"b3c.32x128.encrypted": "HOST_THIRTYTWO_ONEHUNDREDTWENTYEIGHT",
		"m3c.30x240.encrypted": "HOST_THIRTY_TWOHUNDREDFORTY",
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Host Flavor (%s members, %s)", strconv.FormatInt(r.Members, 10), r.Flavor),
		Unit:            "Flavor",
		UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(float64(r.Members))),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr("databases-for-elasticsearch"),
			ProductFamily: strPtr("service"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr(r.Plan),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(CostUnit[r.Flavor]),
		},
	}
}
