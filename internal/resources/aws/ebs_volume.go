package aws

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

var defaultVolumeSize = int64(8)

type EBSVolume struct {
	// "required" args that can't really be missing.
	Address    string
	Region     string
	Type       string
	IOPS       int64
	Throughput int64

	// "optional" args that can be empty strings.
	Size *int64

	// "usage" args
	MonthlyStandardIORequests *int64 `infracost_usage:"monthly_standard_io_requests"`
}

func (a *EBSVolume) CoreType() string {
	return "EBSVolume"
}

func (a *EBSVolume) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_standard_io_requests", DefaultValue: 0, ValueType: schema.Int64},
	}
}

func (a *EBSVolume) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *EBSVolume) BuildResource() *schema.Resource {
	if a.Type == "" {
		a.Type = "gp2"
	}

	costComponents := make([]*schema.CostComponent, 0)
	subResources := make([]*schema.Resource, 0)

	costComponents = append(costComponents, a.storageCostComponent())

	if strings.ToLower(a.Type) == "gp3" && a.Throughput > 125 {
		costComponents = append(costComponents, a.provisionedThroughputCostComponent())
	}

	if strings.ToLower(a.Type) == "io1" {
		costComponents = append(costComponents, a.provisionedIOPSCostComponent("EBS:VolumeP-IOPS.piops", a.IOPS))
	} else if strings.ToLower(a.Type) == "io2" {
		costComponents = append(costComponents, a.provisionedIOPSCostComponent("EBS:VolumeP-IOPS.io2$", a.IOPS))
	} else if strings.ToLower(a.Type) == "gp3" && a.IOPS > 3000 {
		costComponents = append(costComponents, a.provisionedIOPSCostComponent("VolumeP-IOPS.gp3", a.IOPS-3000))
	}

	if strings.ToLower(a.Type) == "standard" {
		costComponents = append(costComponents, a.ioRequestsCostComponent())
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    a.UsageSchema(),
		CostComponents: costComponents,
		SubResources:   subResources,
	}
}

func (a *EBSVolume) storageCostComponent() *schema.CostComponent {
	size := defaultVolumeSize
	if a.Size != nil {
		size = *a.Size
	}

	var name string
	switch strings.ToLower(a.Type) {
	case "standard":
		name = "Storage (magnetic)"
	case "io1":
		name = "Storage (provisioned IOPS SSD, io1)"
	case "io2":
		name = "Storage (provisioned IOPS SSD, io2)"
	case "st1":
		name = "Storage (throughput optimized HDD, st1)"
	case "sc1":
		name = "Storage (cold HDD, sc1)"
	case "gp3":
		name = "Storage (general purpose SSD, gp3)"
	case "gp2":
		name = "Storage (general purpose SSD, gp2)"
	default:
		name = "Storage (unknown)"
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(size)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "volumeApiName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", a.Type))},
			},
		},
	}
}

func (a *EBSVolume) provisionedIOPSCostComponent(usageType string, iops int64) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Provisioned IOPS",
		Unit:            "IOPS",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(iops)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("System Operation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "volumeApiName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", a.Type))},
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
	}
}

func (a *EBSVolume) ioRequestsCostComponent() *schema.CostComponent {
	var qty *decimal.Decimal
	if a.MonthlyStandardIORequests != nil {
		qty = decimalPtr(decimal.NewFromInt(*a.MonthlyStandardIORequests))
	}

	return &schema.CostComponent{
		Name:            "I/O requests",
		Unit:            "1M request",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("System Operation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "volumeApiName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", a.Type))},
				{Key: "usagetype", ValueRegex: strPtr("/EBS:VolumeIOUsage/i")},
			},
		},
		UsageBased: true,
	}
}

func (a *EBSVolume) provisionedThroughputCostComponent() *schema.CostComponent {
	qty := decimal.NewFromInt(a.Throughput - 125)
	qty = qty.Div(decimal.NewFromInt(1024))

	return &schema.CostComponent{
		Name:            "Provisioned throughput",
		Unit:            "Mbps",
		UnitMultiplier:  decimal.NewFromFloat(1.0 / 1024.0),
		MonthlyQuantity: decimalPtr(qty),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(a.Region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("Provisioned Throughput"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "volumeApiName", ValueRegex: strPtr(fmt.Sprintf("/%s/i", a.Type))},
				{Key: "usagetype", ValueRegex: strPtr("/VolumeP-Throughput.gp3/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GiBps-mo"),
		},
	}
}
