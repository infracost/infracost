package google

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

// ComputeGuestAccelerator defines Guest Accelerator setup for Compute resources.
type ComputeGuestAccelerator struct {
	Type  string
	Count int64
}

// ContainerNodeConfig defines Node configuration for Container resources.
type ContainerNodeConfig struct {
	MachineType       string
	PurchaseOption    string
	DiskType          string
	DiskSize          float64
	LocalSSDCount     int64
	GuestAccelerators []*ComputeGuestAccelerator
}

// computeCostComponent returns a cost component for Compute instance usage.
func computeCostComponents(region, machineType string, purchaseOption string, instanceCount int64, monthlyHours *float64) []*schema.CostComponent {
	sustainedUseDiscount := 0.0
	fixPurchaseOption := ""

	if strings.ToLower(purchaseOption) == "on_demand" {
		fixPurchaseOption = "OnDemand"
		switch strings.ToLower(strings.Split(machineType, "-")[0]) {
		case "c2", "n2", "n2d":
			sustainedUseDiscount = 0.2
		case "n1", "f1", "g1", "m1":
			sustainedUseDiscount = 0.3
		}
	}

	purchaseOptionPrefix := ""
	if purchaseOption == "preemptible" {
		purchaseOptionPrefix = "Spot Preemptible "
		fixPurchaseOption = "Preemptible"
	}

	qty := decimal.NewFromFloat(schema.HourToMonthUnitMultiplier.InexactFloat64())
	if monthlyHours != nil {
		qty = decimal.NewFromFloat(*monthlyHours)
	}

	if !strings.Contains(machineType, "custom") {
		return []*schema.CostComponent{
			{
				Name:                fmt.Sprintf("Instance usage (Linux/UNIX, %s, %s)", purchaseOptionLabel(purchaseOption), machineType),
				Unit:                "hours",
				UnitMultiplier:      decimal.NewFromInt(1),
				MonthlyQuantity:     decimalPtr(qty.Mul(decimal.NewFromInt(instanceCount))),
				MonthlyDiscountPerc: sustainedUseDiscount,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Compute Engine"),
					ProductFamily: strPtr("Compute Instance"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "machineType", ValueRegex: regexPtr(fmt.Sprintf("^%s$", machineType))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr(purchaseOption),
				}},
		}
	} else {
		// GCP Custom Instances
		re := regexp.MustCompile(`(\D.+)-(\d+)-(\d.+)`)
		firstMachineTypeInfo := re.ReplaceAllString(machineType, "$1")
		strCPUAmount := re.ReplaceAllString(machineType, "$2")
		strRAMAmount := re.ReplaceAllString(machineType, "$3")

		instanceType := ""
		if firstMachineTypeInfo != "custom" {
			instanceType = strings.ToUpper(strings.Split(firstMachineTypeInfo, "-")[0]) + " "
		}

		ext := " "
		if strings.Contains(strRAMAmount, "ext") {
			ext = " Extended "
			strRAMAmount = strings.Split(strRAMAmount, "-")[0]
		}

		numberOfCores, err := strconv.ParseInt(strCPUAmount, 10, 64)
		if err != nil {
			log.Warnf("Could not parse the custom number of Cores for %s", machineType)
			return nil
		}

		mbOfRAM, err := strconv.ParseInt(strRAMAmount, 10, 64)
		if err != nil {
			log.Warnf("Could not parse the custom amount of RAM for %s", machineType)
			return nil
		}

		cpuCostComponent := &schema.CostComponent{
			Name:                fmt.Sprintf("Custom instance CPU (Linux/UNIX, %s, %s)", purchaseOptionLabel(purchaseOption), machineType),
			Unit:                "hours",
			UnitMultiplier:      decimal.NewFromInt(numberOfCores),
			MonthlyQuantity:     decimalPtr(qty.Mul(decimal.NewFromInt(instanceCount * numberOfCores))),
			MonthlyDiscountPerc: sustainedUseDiscount,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Compute Engine"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: regexPtr(fmt.Sprintf("^%s%sCustom Instance Core", purchaseOptionPrefix, instanceType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr(fixPurchaseOption),
			},
		}

		ramCostComponent := &schema.CostComponent{
			Name:                fmt.Sprintf("Custom Instance RAM (Linux/UNIX, %s, %s)", purchaseOptionLabel(purchaseOption), machineType),
			Unit:                "gibibyte hour",
			UnitMultiplier:      decimal.NewFromInt(mbOfRAM / 1024),
			MonthlyQuantity:     decimalPtr(qty.Mul(decimal.NewFromInt(instanceCount * mbOfRAM / 1024))),
			MonthlyDiscountPerc: sustainedUseDiscount,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Compute Engine"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: regexPtr(fmt.Sprintf("^%s%sCustom%sInstance Ram", purchaseOptionPrefix, instanceType, ext))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr(fixPurchaseOption),
			},
		}

		return []*schema.CostComponent{
			cpuCostComponent,
			ramCostComponent,
		}
	}

}

// bootDiskCostComponent returns a cost component for Boot Disk storage for
// Compute resources.
func bootDiskCostComponent(region string, diskSize float64, diskType string) *schema.CostComponent {
	return computeDiskCostComponent(region, diskType, diskSize, int64(1))
}

// scratchDiskCostComponent returns a cost component for local SSD Disk storage for
// Compute resources.
func scratchDiskCostComponent(region string, purchaseOption string, count int) *schema.CostComponent {
	descRegex := "/^SSD backed Local Storage( in .*)?$/"
	if strings.ToLower(purchaseOption) == "preemptible" {
		descRegex = "/^SSD backed Local Storage attached to Spot Preemptible VMs/"
	}

	return &schema.CostComponent{
		Name:            "Local SSD provisioned storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(375 * count))), // local SSDs are always 375 GB
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(descRegex)},
			},
		},
	}
}

// computeDiskCostComponent returns a cost component for provisioned storage for
// Compute resources.
func computeDiskCostComponent(region string, diskType string, diskSize float64, instanceCount int64) *schema.CostComponent {
	diskTypeDesc := "/^Storage PD Capacity/"
	diskTypeLabel := "Standard provisioned storage (pd-standard)"
	switch diskType {
	case "pd-balanced":
		diskTypeDesc = "/^Balanced PD Capacity/"
		diskTypeLabel = "Balanced provisioned storage (pd-balanced)"
	case "pd-ssd":
		diskTypeDesc = "/^SSD backed PD Capacity/"
		diskTypeLabel = "SSD provisioned storage (pd-ssd)"
	}

	size := decimalPtr(decimal.NewFromInt(instanceCount).Mul(decimal.NewFromFloat(diskSize)))

	return &schema.CostComponent{
		Name:            diskTypeLabel,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: size,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(diskTypeDesc)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""), // use the non-free tier
		},
	}
}

// guestAcceleratorCostComponent returns a cost component for Guest Accelerator usage for Compute resources.
// Callers should be aware guestAcceleratorCostComponent returns nil if the provided guestAcceleratorType is not supported.
func guestAcceleratorCostComponent(region string, purchaseOption string, guestAcceleratorType string, guestAcceleratorCount int64, instanceCount int64, monthlyHours *float64) *schema.CostComponent {
	var (
		name       string
		descPrefix string
	)

	switch guestAcceleratorType {
	case "nvidia-tesla-t4":
		name = "NVIDIA Tesla T4"
		descPrefix = "Nvidia Tesla T4 GPU"
	case "nvidia-tesla-p4":
		name = "NVIDIA Tesla P4"
		descPrefix = "Nvidia Tesla P4 GPU"
	case "nvidia-tesla-v100":
		name = "NVIDIA Tesla V100"
		descPrefix = "Nvidia Tesla V100 GPU"
	case "nvidia-tesla-p100":
		name = "NVIDIA Tesla P100"
		descPrefix = "Nvidia Tesla P100 GPU"
	case "nvidia-tesla-k80":
		name = "NVIDIA Tesla K80"
		descPrefix = "Nvidia Tesla K80 GPU"
	case "nvidia-tesla-a100":
		name = "NVIDIA Tesla A100"
		descPrefix = "Nvidia Tesla A100 GPU"
	default:
		logging.Logger.Debugf("skipping cost component because guest_accelerator.type '%s' is not supported", guestAcceleratorType)
		return nil
	}

	descRegex := fmt.Sprintf("/^%s running/", descPrefix)
	if strings.ToLower(purchaseOption) == "preemptible" {
		descRegex = fmt.Sprintf("/^%s attached to Spot Preemptible VMs running/", descPrefix)
	}

	count := decimal.NewFromInt(guestAcceleratorCount)
	count = decimal.NewFromInt(instanceCount).Mul(count)

	sustainedUseDiscount := 0.0
	if strings.ToLower(purchaseOption) == "on_demand" {
		sustainedUseDiscount = 0.3
	}

	qty := decimal.NewFromFloat(730)
	if monthlyHours != nil {
		qty = decimal.NewFromFloat(*monthlyHours)
	}
	qty = qty.Mul(count)

	return &schema.CostComponent{
		Name:                fmt.Sprintf("%s (%s)", name, purchaseOptionLabel(purchaseOption)),
		Unit:                "hours",
		UnitMultiplier:      decimal.NewFromInt(1),
		MonthlyQuantity:     decimalPtr(qty),
		MonthlyDiscountPerc: sustainedUseDiscount,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(descRegex)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""),
		},
	}
}

// purchaseOptionLabel returns a Purchase Option label based on provider's
// purchase option value.
func purchaseOptionLabel(purchaseOption string) string {
	return map[string]string{
		"on_demand":   "on-demand",
		"preemptible": "preemptible",
	}[purchaseOption]
}

func storageImageCostComponent(region string, description string, storageSize *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: storageSize,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr(description)},
			},
		},
	}
}
