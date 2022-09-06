package ibm

import (
	"encoding/json"
	"fmt"
	"net/http"

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

var IsInstanceUsageSchema = []*schema.UsageItem{
	{Key: "monthly_instance_hours", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the IsInstance.
// It uses the `infracost_usage` struct tags to populate data into the IsInstance.
func (r *IsInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
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
		return "gen2-zvsi-instance"
	default:
		return "unknown"
	}
}

type Metadata struct {
	Other Other `json:"other"`
}

type Other struct {
	Profile Profile `json:"profile"`
}

type Profile struct {
	DefaultConfig DefaultConfig `json:"default_config"`
	Family        string        `json:"family"`
	Generation    string        `json:"generation"`
	ResourceType  string        `json:"resource_type"`
}

type DefaultConfig struct {
	AllowedProfileClasses []string `json:"allowed_profile_classes"`
	Bandwidth             int64    `json:"bandwidth"`
	CPU                   int64    `json:"cpu"`
	Freqency              int64    `json:"freqency"`
	GPUCount              int64    `json:"gpu_count"`
	GPUManufacturer       string   `json:"gpu_manufacturer"`
	GPUMemory             int64    `json:"gpu_memory"`
	GPUModel              string   `json:"gpu_model"`
	OSArchitecture        []string `json:"os_architecture"`
	PortSpeed             int64    `json:"port_speed"`
	RAM                   int64    `json:"ram"`
	VcpuArchitecture      string   `json:"vcpu_architecture"`
	VcpuManufacturer      string   `json:"vcpu_manufacturer"`
	Disks                 []Disk   `json:"disks"`
}

type Disk struct {
	DefaultInterfaceType    string   `json:"default_interface_type"`
	DiskType                string   `json:"disk_type"`
	Quantity                int64    `json:"quantity"`
	Size                    int64    `json:"size"`
	SupportedInterfaceTypes []string `json:"supported_interface_types"`
}

type CatalogInstance struct {
	Id       string   `json:"id"`
	Kind     string   `json:"kind"`
	Metadata Metadata `json:"metadata"`
}

func getProfileFromCatalog(profile string) (CatalogInstance, error) {
	var catalogProfile CatalogInstance
	resp, err := http.Get(fmt.Sprintf("https://globalcatalog.cloud.ibm.com/api/v1/%s?include=metadata", profile))
	if err != nil {
		log.Warn("Request to get instance profile failed", err)
		return catalogProfile, err
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&catalogProfile)
	if err != nil {
		log.Warn(err)
	}
	return catalogProfile, nil
}

func parseArch(arch string) ArchType {
	var archType ArchType
	switch arch {
	case "amd64":
		archType = x86
	case "s390x":
		archType = s390x
	default:
		archType = -1
	}
	return archType
}

func (r *IsInstance) storageCostComponent(arch ArchType, size int64, count int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyInstanceHours))
		quantity = decimalPtr(quantity.Mul(decimal.NewFromInt(size * count)))
	}

	var unit string
	if arch == x86 {
		unit = "IS_STORAGE_GIGABYTE_HOURS"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Storage GB hours (%d GB * %d, %s)", size, count, r.Zone),
		Unit:            "Storage GB hours",
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

func (r *IsInstance) gpuCostComponent(arch ArchType, gpuType string, gpuCount int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyInstanceHours))
		quantity = decimalPtr(quantity.Mul(decimal.NewFromInt(gpuCount)))
	}

	var unit string
	if arch == s390x {
		unit = "V100_HOURS_POWER"
	} else {
		unit = "V100_HOURS"
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Gpu hours (%d GPUs, %s, %s)", gpuCount, gpuType, r.Zone),
		Unit:            "Gpu hours",
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

func (r *IsInstance) memoryCostComponent(arch ArchType, memory int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyInstanceHours))
		quantity = decimalPtr(quantity.Mul(decimal.NewFromInt(memory)))
	}

	var unit = "MEMORY_HOURS"

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
			Unit: strPtr(unit),
		},
	}
}

func (r *IsInstance) cpuCostComponent(arch ArchType, cpu int64) *schema.CostComponent {
	var quantity *decimal.Decimal

	if r.MonthlyInstanceHours != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyInstanceHours))
		quantity = decimalPtr(quantity.Mul(decimal.NewFromInt(cpu)))
	}

	var unit string = "VCPU_HOURS"

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
	var costComponents []*schema.CostComponent

	gcProfile, err := getProfileFromCatalog(r.Profile)

	if err == nil {
		// If the VSI instance runs on a dedicated host, the customer is charged for the dedicated host usages
		if r.IsDedicated {
			costComponents = append(costComponents, r.onDedicatedHostCostComponent(gcProfile.Metadata.Other.Profile.DefaultConfig.CPU, gcProfile.Metadata.Other.Profile.DefaultConfig.RAM))
		} else {
			arch := parseArch(gcProfile.Metadata.Other.Profile.DefaultConfig.VcpuArchitecture)
			costComponents = append(costComponents, r.cpuCostComponent(arch, gcProfile.Metadata.Other.Profile.DefaultConfig.CPU))
			costComponents = append(costComponents, r.memoryCostComponent(arch, gcProfile.Metadata.Other.Profile.DefaultConfig.RAM))
			if gcProfile.Metadata.Other.Profile.DefaultConfig.GPUModel != "" {
				costComponents = append(costComponents, r.gpuCostComponent(arch, gcProfile.Metadata.Other.Profile.DefaultConfig.GPUModel, gcProfile.Metadata.Other.Profile.DefaultConfig.GPUCount))
			}
			if gcProfile.Metadata.Other.Profile.DefaultConfig.Disks != nil {
				for _, disk := range gcProfile.Metadata.Other.Profile.DefaultConfig.Disks {
					costComponents = append(costComponents, r.storageCostComponent(arch, disk.Size, disk.Quantity))
				}
			}
		}
	} else {
		log.Warn(err)
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    IsInstanceUsageSchema,
		CostComponents: costComponents,
	}
}
