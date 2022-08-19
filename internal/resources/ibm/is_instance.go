package ibm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// IsInstance struct represents an IBM virtual server instance.
//
// Pricing information: https://cloud.ibm.com/kubernetes/catalog/about
type IsInstance struct {
	Address              string
	Region               string
	Profile              string // should be values from CLI 'ibmcloud is instance-profiles'
	Zone                 string
	TruncatedZone        string // should be the same as region, but with the last number removed (eg: us-south-1 -> us-south)
	IsDedicated          bool   // will be true if a dedicated_host or dedicated_host_group is specified
	MonthlyInstanceHours *int64 `infracost_usage:"monthly_instance_hours"`
}

type ArchType int64

const (
	x86 ArchType = iota
	s390x
)

func (s ArchType) toPlanName() string {
	switch s {
	case x86:
		return "gen2-instance"
	case s390x:
		return "gen2-instance-power"
	default:
		return "unknown"
	}
}

type instanceInformation struct {
	Dedicated bool
	Storage   bool
	Memory    int64
	Cores     int64
	Arch      ArchType
}

func parseProfile(profile string) instanceInformation {
	parts := strings.Split(profile, "-")
	instanceInfo := instanceInformation{}
	if len(parts) < 2 {
		return instanceInfo
	}
	if len(parts) == 3 && parts[1] == "host" {
		instanceInfo.Dedicated = true
	}
	splitProfileArch := strings.Split(parts[0], "")
	if len(splitProfileArch) > 2 {
		switch splitProfileArch[1] {
		case "x":
			instanceInfo.Arch = x86
		case "z":
			instanceInfo.Arch = s390x
		default:
			break
		}
	}

	cpuAndMemory := strings.Split(parts[len(parts)-1], "x")
	if len(cpuAndMemory) == 2 {
		var err error
		instanceInfo.Cores, err = strconv.ParseInt(cpuAndMemory[0], 10, 64)
		if err != nil {
			log.Warnf("Parse error at is.instance profile parsing: %s", profile)
		}
		instanceInfo.Memory, err = strconv.ParseInt(cpuAndMemory[1], 10, 64)
		if err != nil {
			log.Warnf("Parse error at is.instance profile parsing: %s", profile)
		}
	}
	return instanceInfo
}

var IsInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsInstance.
// It uses the `infracost_usage` struct tags to populate data into the IsInstance.
func (r *IsInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *IsInstance) memoryCostComponent(arch ArchType, memory int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyInstanceHours))
		quantity = decimalPtr(quantity.Mul(decimal.NewFromInt(memory)))
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Memory hours (%d GB, %s)", memory, r.Zone),
		Unit:            "Memory hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr(arch.toPlanName())},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("MEMORY_HOURS"),
		},
	}
}

func (r *IsInstance) cpuCostComponent(arch ArchType, cpu int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyInstanceHours))
		quantity = decimalPtr(quantity.Mul(decimal.NewFromInt(cpu)))
	}

	var unit string
	if arch == s390x {
		unit = "VCPU_HOURS_POWER"
	} else {
		unit = "VCPU_HOURS"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("CPU hours (%d CPUs, %s)", cpu, r.Zone),
		Unit:            "CPU hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "planName", Value: strPtr(arch.toPlanName())},
			},
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr(unit),
		},
	}
}

func (r *IsInstance) onDedicatedHostCostComponent(cores int64, memory int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyInstanceHours))
	}
	costCompoment := &schema.CostComponent{
		Name:            fmt.Sprintf("Host Hours (%d vCPUs, %d GB, %s)", cores, memory, r.Zone),
		Unit:            "hours",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("ibm"),
			Region:     strPtr(r.Region),
			Service:    strPtr("is.instance"),
		},
	}
	costCompoment.SetCustomPrice(decimalPtr(decimal.NewFromInt(0)))
	return costCompoment
}

// BuildResource builds a schema.Resource from a valid IsInstance struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *IsInstance) BuildResource() *schema.Resource {
	profileInfo := parseProfile(r.Profile)

	var costComponents []*schema.CostComponent

	// If the VSI instance runs on a dedicated host, the customer is charged for the dedicated host usages
	if r.IsDedicated {
		costComponents = append(costComponents, r.onDedicatedHostCostComponent(profileInfo.Cores, profileInfo.Memory))
	} else {
		costComponents = append(costComponents, r.cpuCostComponent(profileInfo.Arch, profileInfo.Cores))
		costComponents = append(costComponents, r.memoryCostComponent(profileInfo.Arch, profileInfo.Memory))
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsInstanceUsageSchema,
		CostComponents: costComponents,
	}
}
