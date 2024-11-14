package ibm

import (
	"math"
	"fmt"
	"strconv"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

const MILLION_HTTP_REQUESTS float64 = 1000000

// CodeEngineFunction struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://cloud.ibm.com/docs/codeengine?topic=codeengine-getting-started
// Pricing information: https://cloud.ibm.com/docs/codeengine?topic=codeengine-pricing
type CodeEngineFunction struct {
	Address string
	Region  string
	CPU string
	Memory string

	HttpRequestCalls *float64 `infracost_usage:"http_request_calls"`
	InstanceHours *float64 `infracost_usage:"instance_hours"`
}

// CodeEngineFunctionUsageSchema defines a list which represents the usage schema of CodeEngineFunction.
var CodeEngineFunctionUsageSchema = []*schema.UsageItem{
	{Key: "http_request_calls", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "instance_hours", DefaultValue: 1, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the CodeEngineFunction.
// It uses the `infracost_usage` struct tags to populate data into the CodeEngineFunction.
func (r *CodeEngineFunction) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CodeEngineFunction) CodeEngineFunctionVirtualProcessorCoreCostComponent() *schema.CostComponent {
	var q, err = strconv.ParseFloat(r.CPU, 64)
	if err != nil {
		q = float64(1) // Default 1 vCPU
	}

	var hours *decimal.Decimal
	if (r.InstanceHours != nil) {
		hours = decimalPtr(decimal.NewFromFloat(*r.InstanceHours * q))
	}
	
	return &schema.CostComponent{
		Name:			fmt.Sprintf("Virtual Processor Cores"),
		Unit:			"vCPU Hours",
		UnitMultiplier:	decimal.NewFromInt(100),
		MonthlyQuantity: hours,
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

func (r *CodeEngineFunction) CodeEngineFunctionRAMCostComponent() *schema.CostComponent {
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

	var hours *decimal.Decimal
	if (r.InstanceHours != nil) {
		hours = decimalPtr(decimal.NewFromFloat(*r.InstanceHours * memGB))
	}
	
	return &schema.CostComponent{
		Name:			fmt.Sprintf("RAM"),
		Unit:			"GB Hours",
		UnitMultiplier:	decimal.NewFromInt(100),
		MonthlyQuantity: hours,
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

func (r *CodeEngineFunction) CodeEngineFunctionHTTPRequestsCostComponent() *schema.CostComponent {
	var q *decimal.Decimal
	if r.HttpRequestCalls != nil {
		mil_req := math.Ceil(*r.HttpRequestCalls / MILLION_HTTP_REQUESTS)
		q = decimalPtr(decimal.NewFromFloat(mil_req))
	}

	return &schema.CostComponent{
		Name:			"Million HTTP calls",
		Unit:			"Million HTTP calls",
		UnitMultiplier:	decimal.NewFromInt(5),
		MonthlyQuantity: q,
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
			Unit: strPtr("MILLION_API_CALLS"),
		},
	}
}

// BuildResource builds a schema.Resource from a valid CodeEngineFunction struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CodeEngineFunction) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.CodeEngineFunctionVirtualProcessorCoreCostComponent(),
		r.CodeEngineFunctionRAMCostComponent(),
		r.CodeEngineFunctionHTTPRequestsCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    CodeEngineFunctionUsageSchema,
		CostComponents: costComponents,
	}
}
