package aws

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type MWAAEnvironment struct {
	// These fields are required since they are pulled directly from the IAC configuration (e.g. the terraform plan)
	Address string
	Region  string
	Size    string // Should be Small, Medium, or Large

	// If there is a parameter than needs to be read from infracost-usage.yml you define it like this:
	AdditionalWorkers    *float64 `infracost_usage:"additional_workers"`
	AdditionalSchedulers *float64 `infracost_usage:"additional_schedulers"`
	MetaDatabaseGB       *float64 `infracost_usage:"meta_database_gb"`
}

// If the resource requires a usage parameter
func (a *MWAAEnvironment) CoreType() string {
	return "MWAAEnvironment"
}

func (a *MWAAEnvironment) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "additional_workers", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "additional_schedulers", DefaultValue: 0.0, ValueType: schema.Float64},
		{Key: "meta_database_gb", DefaultValue: 0.0, ValueType: schema.Float64},
	}
}

// If the resource requires a usage parameter
func (a *MWAAEnvironment) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(a, u)
}

func (a *MWAAEnvironment) BuildResource() *schema.Resource {
	var workerQuantity, schedulerQuantity, metaDatabaseGB *decimal.Decimal
	if a.AdditionalWorkers != nil {
		workerQuantity = decimalPtr(decimal.NewFromFloat(*a.AdditionalWorkers))
	}
	if a.AdditionalSchedulers != nil {
		schedulerQuantity = decimalPtr(decimal.NewFromFloat(*a.AdditionalSchedulers))
	}
	if a.MetaDatabaseGB != nil {
		metaDatabaseGB = decimalPtr(decimal.NewFromFloat(*a.MetaDatabaseGB))
	}

	costComponents := []*schema.CostComponent{
		a.newInstanceCostComponent("Environment", a.Size, decimalPtr(decimal.NewFromInt(1))),
		a.newInstanceCostComponent("Worker", a.Size, workerQuantity),
		a.newInstanceCostComponent("Scheduler", a.Size, schedulerQuantity),
		a.newStorageCostComponent(metaDatabaseGB),
	}

	return &schema.Resource{
		Name:           a.Address,
		UsageSchema:    a.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (a *MWAAEnvironment) newInstanceCostComponent(instanceType, size string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s (%s)", instanceType, size),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(a.Region),
			Service:    strPtr("AmazonMWAA"),
			AttributeFilters: []*schema.AttributeFilter{
				// Note the use of start/end anchors and case-insensitive match with ValueRegex
				{Key: "size", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", a.Size))},
				{Key: "type", Value: strPtr(instanceType)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (a *MWAAEnvironment) newStorageCostComponent(quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Meta database",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(a.Region),
			Service:    strPtr("AmazonMWAA"),
			AttributeFilters: []*schema.AttributeFilter{
				// Note the use of start/end anchors and case-insensitive match with ValueRegex
				{Key: "usagetype", ValueRegex: strPtr("/Airflow-StandardDatabaseStorage$/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}
