package ibm

import (
	"fmt"
	"strconv"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// CodeEngineJob struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://cloud.ibm.com/docs/codeengine?topic=codeengine-getting-started
// Pricing information: https://cloud.ibm.com/docs/codeengine?topic=codeengine-pricing
type CodeEngineJob struct {
	Address string
	Region  string
	CPU string
	Memory string

	InstanceHours *float64 `infracost_usage:"instance_hours"`
	ScaledInstances *float64 `infracost_usage:"scaled_instances"`
}

// CodeEngineJobUsageSchema defines a list which represents the usage schema of CodeEngineJob.
var CodeEngineJobUsageSchema = []*schema.UsageItem{
	{Key: "instance_hours", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "scaled_instances", DefaultValue: 1, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the CodeEngineJob.
// It uses the `infracost_usage` struct tags to populate data into the CodeEngineJob.
func (r *CodeEngineJob) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}


func (r *CodeEngineJob) CodeEngineJobVirtualProcessorCoreCostComponent() *schema.CostComponent {
	var q, err = strconv.ParseFloat(r.CPU, 64)
	if err != nil {
		q = float64(1) // Default 1 vCPU
	}

	var totalvcpu float64
	if r.ScaledInstances != nil {
		totalvcpu = q * *r.ScaledInstances
	}

	var ih *decimal.Decimal
	if r.InstanceHours != nil {
		ih = decimalPtr(decimal.NewFromFloat(*r.InstanceHours * totalvcpu))
	}

	return &schema.CostComponent{
		Name:			fmt.Sprintf("Virtual Processor Cores Hours"),
		Unit:			"vCPU Hours",
		UnitMultiplier:	decimal.NewFromInt(100),
		MonthlyQuantity: ih,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region: 	strPtr(r.Region),
			Service: 	strPtr("codeengine"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr("standard"),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("VIRTUAL_PROCESSOR_CORE_HOURS"),
		},
	}
}

func (r *CodeEngineJob) CodeEngineJobRAMCostComponent() *schema.CostComponent {
	var memGB float64
	if r.Memory != "" {
		trimmedMemory := r.Memory[:len(r.Memory)-1]
		memGB, _ = strconv.ParseFloat(trimmedMemory, 64)
		if string(r.Memory[len(r.Memory)-1]) == "M" {
			memGB = memGB / float64(1024)
		}
	} else {
		memGB = float64(4) // Default 4GB
	}

	var totalmemGB float64
	if r.ScaledInstances != nil {
		totalmemGB = memGB * *r.ScaledInstances
	}

	var ih *decimal.Decimal
	if r.InstanceHours != nil {
		ih = decimalPtr(decimal.NewFromFloat(*r.InstanceHours * totalmemGB))
	}

	return &schema.CostComponent{
		Name:			fmt.Sprintf("RAM Hours"),
		Unit:			"GB Hours",
		UnitMultiplier:	decimal.NewFromInt(100),
		MonthlyQuantity: ih,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region: 	strPtr(r.Region),
			Service: 	strPtr("codeengine"),
			AttributeFilters: []*schema.AttributeFilter{
				{
					Key: "planName", Value: strPtr("standard"),
				},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_HOURS"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid CodeEngineJob struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CodeEngineJob) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.CodeEngineJobVirtualProcessorCoreCostComponent(),
		r.CodeEngineJobRAMCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    CodeEngineJobUsageSchema,
		CostComponents: costComponents,
	}
}
