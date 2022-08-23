package ibm

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// PiInstance struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://cloud.ibm.com/<PATH/TO/RESOURCE>/
// Pricing information: https://cloud.ibm.com/<PATH/TO/PRICING>/

type PiInstance struct {
	Address         string
	Region          string
	ProcessorMode   string
	SystemType      string
	StorageType     string
	OperatingSystem int64
	Memory          float64
	Cpus            float64

	Storage *float64 `infracost_usage:"storage"`
}

// Operating System
const (
	AIX int64 = iota
	IBMI
	RHEL
	SLES
)

const s922 string = "s922"
const e980 string = "e980"
const e1080 string = "e1080"

// PiInstanceUsageSchema defines a list which represents the usage schema of PiInstance.
var PiInstanceUsageSchema = []*schema.UsageItem{
	{Key: "storage", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the PiInstance.
// It uses the `infracost_usage` struct tags to populate data into the PiInstance.
func (r *PiInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid PiInstance struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *PiInstance) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.piInstanceCoresCostComponent(),
		r.piInstanceMemoryCostComponent(),
		r.piInstanceStorageCostComponent(),
		r.piInstanceOperatingSystemCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    PiInstanceUsageSchema,
		CostComponents: costComponents,
	}
}

func (r *PiInstance) piInstanceOperatingSystemCostComponent() *schema.CostComponent {
	unit := ""

	if r.OperatingSystem == AIX {
		if r.SystemType == s922 {
			unit = "AIX_SMALL_APPLICATION_INSTANCE_HOURS"
		} else if r.SystemType == e980 || r.SystemType == e1080 {
			unit = "AIX_MEDIUM_APPLICATION_INSTANCE_HOURS"
		}
	} else if r.OperatingSystem == IBMI {
		if r.SystemType == s922 {
			unit = ""
		}
	} else if r.OperatingSystem == RHEL {
		unit = ""
	} else if r.OperatingSystem == SLES {
		unit = ""
	}

	return &schema.CostComponent{
		Name:           "Operating System",
		Unit:           "Instance",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("service"),
			Service:       strPtr("power-iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr("power-virtual-server-group")},
				{Key: "planType", Value: strPtr("Paid")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(unit),
		},
	}
}

func (r *PiInstance) piInstanceCoresCostComponent() *schema.CostComponent {
	q := decimalPtr(decimal.NewFromFloat(r.Cpus))

	unit := ""

	if r.ProcessorMode == "shared" {
		if r.SystemType == s922 {
			unit = "SOS_VIRTUAL_PROCESSOR_CORE_HOURS"
		} else if r.SystemType == e980 {
			unit = "ESS_VIRTUAL_PROCESSOR_CORE_HOURS"
		} else if r.SystemType == e1080 {
			unit = "PTEN_ESS_VIRTUAL_PROCESSOR_CORE_HRS"
		}
	} else if r.ProcessorMode == "dedicated" {
		if r.SystemType == s922 {
			unit = "SOD_VIRTUAL_PROCESSOR_CORE_HOURS"
		} else if r.SystemType == e980 {
			unit = "EDD_VIRTUAL_PROCESSOR_CORE_HOURS"
		} else if r.SystemType == e1080 {
			unit = "PTEN_EDD_VIRTUAL_PROCESSOR_CORE_HRS"
		}
	} else if r.ProcessorMode == "capped" {
		if r.SystemType == s922 {
			unit = "SOC_VIRTUAL_PROCESSOR_CORE_HOURS"
		} else if r.SystemType == e980 {
			unit = "ECC_VIRTUAL_PROCESSOR_CORE_HOURS"
		} else if r.SystemType == e1080 {
			unit = "PTEN_ECC_VIRTUAL_PROCESSOR_CORE_HRS"
		}
	}

	return &schema.CostComponent{
		Name:           "Cores",
		Unit:           "Core",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("service"),
			Service:       strPtr("power-iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr("power-virtual-server-group")},
				{Key: "planType", Value: strPtr("Paid")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(unit),
		},
	}
}

func (r *PiInstance) piInstanceMemoryCostComponent() *schema.CostComponent {
	q := decimalPtr(decimal.NewFromFloat(r.Memory))

	unit := "MS_GIGABYTE_HOURS"

	return &schema.CostComponent{
		Name:           "Memory",
		Unit:           "GB",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("service"),
			Service:       strPtr("power-iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr("power-virtual-server-group")},
				{Key: "planType", Value: strPtr("Paid")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(unit),
		},
	}
}

func (r *PiInstance) piInstanceStorageCostComponent() *schema.CostComponent {

	var q *decimal.Decimal

	if r.Storage != nil {
		q = decimalPtr(decimal.NewFromFloat(*r.Storage))
	}

	unit := ""

	if r.StorageType == "tier1" {
		unit = "TIER_ONE_STORAGE_GIGABYTE_HOURS"
	} else if r.StorageType == "tier3" {
		unit = "TIER_THREE_STORAGE_GIGABYTE_HOURS"
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("Storage - %s", r.StorageType),
		Unit:           "GB",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("service"),
			Service:       strPtr("power-iaas"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr("power-virtual-server-group")},
				{Key: "planType", Value: strPtr("Paid")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(unit),
		},
	}
}
