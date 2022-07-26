package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

// ResourceInstance struct represents a resource instance
//
// This terraform resource is opaque and can handle a number of services, provided with the right parameters
type ResourceInstance struct {
	Address    string
	Service    string
	Plan       string
	Location   string
	Parameters gjson.Result

	// KMS
	KMS_ItemsPerMonth *int64 `infracost_usage:"kms_items_per_month"`
	// Secrets Manager
}

// ResourceInstanceUsageSchema defines a list which represents the usage schema of ResourceInstance.
var ResourceInstanceUsageSchema = []*schema.UsageItem{
	{Key: "kms_items_per_month", DefaultValue: 0, ValueType: schema.Int64},
}

// PopulateUsage parses the u schema.UsageData into the ResourceInstance.
// It uses the `infracost_usage` struct tags to populate data into the ResourceInstance.
func (r *ResourceInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

type ResourceCostComponents []*schema.CostComponent

func KMSFreeCostComponent(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.KMS_ItemsPerMonth != nil {
		q = decimalPtr(decimal.NewFromInt(*r.KMS_ItemsPerMonth))
	}
	if q.GreaterThan(decimal.NewFromInt(20)) {
		q = decimalPtr(decimal.NewFromInt(20))
	}
	costComponent := schema.CostComponent{
		Name:            fmt.Sprintf("Items free allowance (first 20 Items)"),
		Unit:            "Item",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    strPtr("kms"),
		},
	}
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return &costComponent
}

func KMSTierCostComponents(r *ResourceInstance) *schema.CostComponent {
	var q *decimal.Decimal
	if r.KMS_ItemsPerMonth != nil {
		q = decimalPtr(decimal.NewFromInt(*r.KMS_ItemsPerMonth))
	}
	if q.LessThanOrEqual(decimal.NewFromInt(20)) {
		q = decimalPtr(decimal.NewFromInt(0))
	} else {
		q = decimalPtr(q.Sub(decimal.NewFromInt(20)))
	}
	costComponent := schema.CostComponent{
		Name:            fmt.Sprintf("Items"),
		Unit:            "Item",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
	}
	return &costComponent
}

func GetKMSCostComponents(r *ResourceInstance) []*schema.CostComponent {
	return []*schema.CostComponent{
		KMSFreeCostComponent(r),
		KMSTierCostComponents(r),
	}
}

// BuildResource builds a schema.Resource from a valid ResourceInstance struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ResourceInstance) BuildResource() *schema.Resource {
	resourceCostMap := make(map[string]ResourceCostComponents)
	resourceCostMap["kms"] = GetKMSCostComponents(r)

	costComponents := resourceCostMap[r.Service]

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    ResourceInstanceUsageSchema,
		CostComponents: costComponents,
	}
}
