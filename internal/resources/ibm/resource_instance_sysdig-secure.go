package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

const SCCWP_GRADUATED_PLAN_PROGRAMMATIC_NAME string = "graduated-tier"
const SCCWP_TRIAL_PLAN_PROGRAMMATIC_NAME string = "free-trial"

func GetSCCWPCostComponents(r *ResourceInstance) []*schema.CostComponent {
	if r.Plan == SCCWP_TRIAL_PLAN_PROGRAMMATIC_NAME {
		costComponent := schema.CostComponent{
			Name:            "Free Trial",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	} else if r.Plan == SCCWP_GRADUATED_PLAN_PROGRAMMATIC_NAME {
		return []*schema.CostComponent{
			SCCWPMultiCloudCSPMComputeInstancesCostComponent(r),
			SCCWPNodeHoursCostComponent(r),
			SCCWPVMNodeHoursCostComponent(r),
		}
	} else {
		costComponent := schema.CostComponent{
			Name:            fmt.Sprintf("Plan %s with customized pricing", r.Plan),
			UnitMultiplier:  decimal.NewFromInt(1), // Final quantity for this cost component will be divided by this amount
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
		return []*schema.CostComponent{
			&costComponent,
		}
	}
}

/*
 * Graduated Tier Plan: 6 tiers
 */
func SCCWPMultiCloudCSPMComputeInstancesCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal
	if r.SCCWP_MulticloudCSPMComputeInstances != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.SCCWP_MulticloudCSPMComputeInstances))
	}

	costComponent := schema.CostComponent{
		Name:            "Multi-Cloud CSPM Compute Instance Hours",
		Unit:            "Instance-Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("MULTI_CLOUD_CSPM_COMPUTE_INSTANCES"),
		},
	}
	return &costComponent
}

/*
 * Graduated Tier Plan: 6 tiers
 */
func SCCWPNodeHoursCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal
	if r.SCCWP_NodeHours != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.SCCWP_NodeHours))
	}

	costComponent := schema.CostComponent{
		Name:            "Node Hours",
		Unit:            "Instance-Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("NODE_HOURS"),
		},
	}
	return &costComponent
}

/*
 * Graduated Tier Plan: 6 tiers
 */
func SCCWPVMNodeHoursCostComponent(r *ResourceInstance) *schema.CostComponent {

	var quantity *decimal.Decimal
	if r.SCCWP_VMNodeHours != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.SCCWP_VMNodeHours))
	}

	costComponent := schema.CostComponent{
		Name:            "VM Node Hours",
		Unit:            "Instance-Hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Location),
			Service:    &r.Service,
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: &r.Plan},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("VM_NODE_HOUR"),
		},
	}
	return &costComponent
}
