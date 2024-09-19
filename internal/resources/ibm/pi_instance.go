package ibm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// PiInstance struct represents a Virtual Power Systems instance
//
// Resource information: https://www.ibm.com/products/power-virtual-server
// Pricing information: https://cloud.ibm.com/catalog/services/power-systems-virtual-server
// Detailed pricing information: https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-pricing-virtual-server

type PiInstance struct {
	Address                string
	Region                 string
	ProcessorMode          string
	SystemType             string
	StorageType            string
	OperatingSystem        int64
	Memory                 float64
	Cpus                   float64
	LegacyIBMiImageVersion bool
	NetweaverImage         bool
	Profile                string

	MonthlyInstanceHours      *float64 `infracost_usage:"monthly_instance_hours"`
	Storage                   *float64 `infracost_usage:"storage"`
	CloudStorageSolution      *int64   `infracost_usage:"cloud_storage_solution"`
	HighAvailability          *int64   `infracost_usage:"high_availability"`
	DB2WebQuery               *int64   `infracost_usage:"db2_web_query"`
	RationalDevStudioLicences *int64   `infracost_usage:"rational_dev_studio_licenses"`
	Epic                      *int64   `infracost_usage:"epic"`
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
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "storage", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "cloud_storage_solution", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "high_availability", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "db2_web_query", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "rational_dev_studio_licenses", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "epic", DefaultValue: 0, ValueType: schema.Int64},
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
		r.piInstanceStorageCostComponent(),
	}

	if r.Profile != "" {
		costComponents = append(costComponents, r.piInstanceMemoryHanaProfileCostComponent(), r.piInstanceCoresHanaProfileCostComponent())
	} else {
		costComponents = append(costComponents, r.piInstanceCoresCostComponent(), r.piInstanceMemoryCostComponent())
	}

	if r.OperatingSystem == AIX {
		costComponents = append(costComponents, r.piInstanceAIXOperatingSystemCostComponent())
	} else if r.OperatingSystem == IBMI {
		costComponents = append(costComponents,
			r.piInstanceIBMiLPPPOperatingSystemCostComponent(),
			r.piInstanceIBMiOSOperatingSystemCostComponent(),
			r.piInstanceCloudStorageSolutionCostComponent(),
			r.piInstanceHighAvailabilityCostComponent(),
			r.piInstanceDB2WebQueryCostComponent(),
			r.piInstanceRationalDevStudioLicensesCostComponent(),
		)
		if r.LegacyIBMiImageVersion {
			costComponents = append(costComponents, r.piInstanceIBMiOperatingSystemServiceExtensionCostComponent())
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    PiInstanceUsageSchema,
		CostComponents: costComponents,
	}
}

func (r *PiInstance) piInstanceAIXOperatingSystemCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Cpus * hours))
	}

	unit := ""

	if r.SystemType == s922 {
		unit = "AIX_SMALL_APPLICATION_INSTANCE_HOURS"
	} else if r.SystemType == e980 || r.SystemType == e1080 {
		unit = "AIX_MEDIUM_APPLICATION_INSTANCE_HOURS"
	}

	return &schema.CostComponent{
		Name:            "Operating System",
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceIBMiLPPPOperatingSystemCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Cpus * hours))
	}

	unit := ""

	if r.SystemType == s922 {
		unit = "IBMI_LPP_PTEN_APPLICATION_INSTANCE_HOURS"
	} else if r.SystemType == e980 {
		unit = "IBMI_LPP_PTHIRTY_APPLICATION_INSTANCE_HOURS"
	}

	return &schema.CostComponent{
		Name:            "Operating System IBMi LPP",
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceIBMiOSOperatingSystemCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Cpus * hours))
	}

	unit := ""

	if r.SystemType == s922 {
		unit = "IBMI_OS_PTEN_APPLICATION_INSTANCE_HOURS"
	} else if r.SystemType == e980 {
		unit = "IBMI_OS_PTHIRTY_APPLICATION_INSTANCE_HOURS"
	}

	return &schema.CostComponent{
		Name:            "Operating System IBMi OS",
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceIBMiOperatingSystemServiceExtensionCostComponent() *schema.CostComponent {
	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Cpus * hours))
	}

	unit := "IBM_I_OS_PTEN_SRVC_EXT_PER_PROC_CORE_HR"

	if r.SystemType == e980 {
		unit = "IBM_I_SERVICE_EXTENSION_PER_CORE_HOUR"
	}

	return &schema.CostComponent{
		Name:            "Operating System IBMi Service Extension",
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceMemoryHanaProfileCostComponent() *schema.CostComponent {
	var memoryAmount float64

	if r.Profile != "" {
		coresAndMemory := strings.Split(r.Profile, "-")[1]
		memoryString := strings.Split(coresAndMemory, "x")[1]
		memory, err := strconv.Atoi(memoryString)
		if err == nil {
			memoryAmount = float64(memory)
		}
	}

	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(memoryAmount * hours))
	}

	unit := "MEMHANA_APPLICATION_INSTANCE_HOURS"

	return &schema.CostComponent{
		Name:            "Linux HANA Memory",
		Unit:            "Memory hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceCoresHanaProfileCostComponent() *schema.CostComponent {
	var coresAmount float64

	if r.Profile != "" {
		coresAndMemory := strings.Split(r.Profile, "-")[1]
		coresString := strings.Split(coresAndMemory, "x")[0]
		cores, err := strconv.Atoi(coresString)
		if err == nil {
			coresAmount = float64(cores)
		}
	}

	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(coresAmount * hours))
	}

	unit := "COREHANA_APPLICATION_INSTANCE_HOURS"

	return &schema.CostComponent{
		Name:            "Linux HANA Cores",
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceCloudStorageSolutionCostComponent() *schema.CostComponent {
	var cloudStorageSolutionAmount float64
	var q *decimal.Decimal

	if r.CloudStorageSolution != nil {
		cloudStorageSolutionAmount = float64(*r.CloudStorageSolution)
	}

	unit := "IBMI_CSS_APPLICATION_INSTANCE_HOURS"

	if r.MonthlyInstanceHours != nil && r.CloudStorageSolution != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Cpus * cloudStorageSolutionAmount * hours))
	}

	return &schema.CostComponent{
		Name:            "Cloud Storage Solution",
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceHighAvailabilityCostComponent() *schema.CostComponent {
	var highAvailabilityAmount float64

	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil && r.HighAvailability != nil {
		highAvailabilityAmount = float64(*r.HighAvailability)
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Cpus * highAvailabilityAmount * hours))
	}

	unit := "IBMIHA_PTHIRTY_APPLICATION_INSTANCES"

	return &schema.CostComponent{
		Name:            "High Availability",
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceDB2WebQueryCostComponent() *schema.CostComponent {
	var db2WebQueryAmount float64
	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil && r.DB2WebQuery != nil {
		db2WebQueryAmount = float64(*r.DB2WebQuery)
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Cpus * db2WebQueryAmount * hours))
	}

	unit := "IBMI_DBIIWQ_APPLICATION_INSTANCE_HOURS"

	return &schema.CostComponent{
		Name:            "IBM DB2 Web Query",
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

func (r *PiInstance) piInstanceRationalDevStudioLicensesCostComponent() *schema.CostComponent {
	var rationalDevStudioLicencesAmount float64
	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil && r.RationalDevStudioLicences != nil {
		rationalDevStudioLicencesAmount = float64(*r.RationalDevStudioLicences)
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(rationalDevStudioLicencesAmount * hours))
	}

	unit := "IBMIRDS_APPLICATION_INSTANCES"

	return &schema.CostComponent{
		Name:            "Rational Dev Studio",
		Unit:            "Instance hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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
	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Cpus * hours))
	}

	epicEnabled := r.Epic != nil && *r.Epic == 1

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
			if epicEnabled {
				unit = "ESS_VIRTUAL_PROCESSOR_CORE_HOURS"
			} else {
				if r.OperatingSystem == SLES && r.Profile != "" {
					unit = "COREHANA_APPLICATION_INSTANCE_HOURS"
				} else {
					unit = "EDD_VIRTUAL_PROCESSOR_CORE_HOURS"
				}
			}
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

	name := "Cores"

	if epicEnabled {
		name = "Cores - Epic enabled"
	}

	return &schema.CostComponent{
		Name:            name,
		Unit:            "Core hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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
	var q *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(r.Memory * hours))
	}

	unit := "MS_GIGABYTE_HOURS"

	if r.OperatingSystem == SLES && !r.NetweaverImage {
		unit = "MEMHANA_APPLICATION_INSTANCE_HOURS"
	}

	return &schema.CostComponent{
		Name:            "Memory",
		Unit:            "GB hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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

	if r.Storage != nil && r.MonthlyInstanceHours != nil {
		hours := *r.MonthlyInstanceHours
		q = decimalPtr(decimal.NewFromFloat(*r.Storage * hours))
	}

	unit := ""

	if r.StorageType == "tier1" {
		unit = "TIER_ONE_STORAGE_GIGABYTE_HOURS"
	} else if r.StorageType == "tier3" {
		unit = "TIER_THREE_STORAGE_GIGABYTE_HOURS"
	} else if r.StorageType == "tier0" {
		unit = "TIER_ZERO_STORAGE_GIGABYTE_HOURS"
	} else if r.StorageType == "tier5k" {
		unit = "FIXED_5K_OPS_GIGABYTE_HOURS"
	}
	return &schema.CostComponent{
		Name:            fmt.Sprintf("Storage - %s", r.StorageType),
		Unit:            "GB hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: q,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Region),
			ProductFamily: strPtr("composite"),
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
