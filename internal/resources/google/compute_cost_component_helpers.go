package google

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

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

/*
For the values initialized in the struct below, please refer to the

	sustained-use-discounts section of the GCP Documentation.

size_SUD is set to 4 as the Usage level (% of month) is 4

	to make sure we don't overshoot the array
*/
const sudRateSize = 4

type sudRates struct {
	thresholds [sudRateSize]float64
	rates      [sudRateSize]float64
}

var sudRate20 = sudRates{
	thresholds: [sudRateSize]float64{0.25, 0.50, 0.75, 1.0},
	rates:      [sudRateSize]float64{1.0, 0.8678, 0.733, 0.6},
}

var sudRate30 = sudRates{
	thresholds: [sudRateSize]float64{0.25, 0.50, 0.75, 1.0},
	rates:      [sudRateSize]float64{1.0, 0.80, 0.60, 0.40},
}

// computeCostComponent returns a cost component for Compute instance usage.
func computeCostComponents(region, machineType string, purchaseOption string, instanceCount int64, monthlyHours *float64) ([]*schema.CostComponent, error) {
	if strings.HasPrefix(strings.ToLower(machineType), "e2-custom") {
		return nil, errors.New("Infracost currently does not support E2 custom instances")
	}

	sustainedUseDiscount := 0.0
	fixPurchaseOption := ""
	hours, _ := schema.HourToMonthUnitMultiplier.Float64()

	if monthlyHours != nil {
		// Assume 730 hours(max) if monthly_hrs is not set
		if *monthlyHours == 0 {
			*monthlyHours = 730.0
		}
		hours = *monthlyHours
	}

	if strings.ToLower(purchaseOption) == "on_demand" {
		fixPurchaseOption = "OnDemand"
		switch strings.ToLower(strings.Split(machineType, "-")[0]) {
		case "c2", "n2", "n2d":
			sustainedUseDiscount = getSustainedUseDiscount(hours, sudRate20)
		case "custom", "n1", "f1", "g1", "m1":
			sustainedUseDiscount = getSustainedUseDiscount(hours, sudRate30)
		default:
			sustainedUseDiscount = 0.0
		}
	}

	purchaseOptionPrefix := ""
	if purchaseOption == "preemptible" {
		purchaseOptionPrefix = "Spot Preemptible "
		fixPurchaseOption = "Preemptible"
	}

	qty := schema.HourToMonthUnitMultiplier
	if monthlyHours != nil {
		qty = decimal.NewFromFloat(*monthlyHours)
	}
	qty = qty.Mul(decimal.NewFromInt(instanceCount))

	if !strings.Contains(machineType, "custom") {
		return []*schema.CostComponent{
			{
				Name:                fmt.Sprintf("Instance usage (Linux/UNIX, %s, %s)", purchaseOptionLabel(purchaseOption), machineType),
				Unit:                "hours",
				UnitMultiplier:      decimal.NewFromInt(1),
				MonthlyQuantity:     decimalPtr(qty),
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
		}, nil
	} else {
		// GCP Custom Instances
		re := regexp.MustCompile(`(\D.+)-(\d+)-(\d.+)`)
		machineTypePrefix := re.ReplaceAllString(machineType, "$1")
		strCPUAmount := re.ReplaceAllString(machineType, "$2")
		strRAMAmount := re.ReplaceAllString(machineType, "$3")

		extended := false
		if strings.Contains(strRAMAmount, "ext") {
			extended = true
			strRAMAmount = strings.Split(strRAMAmount, "-")[0]
		}

		instanceType := "N1"
		instanceTypePrefix := "Custom"

		if machineTypePrefix != "custom" {
			instanceType = strings.ToUpper(strings.Split(machineTypePrefix, "-")[0])

			if strings.HasSuffix(instanceType, "D") {
				instanceTypePrefix = fmt.Sprintf("%s AMD Custom", instanceType)
			} else if instanceType != "N1" {
				instanceTypePrefix = fmt.Sprintf("%s Custom", instanceType)
			}
		}

		cores, err := strconv.ParseInt(strCPUAmount, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse the custom number of cores for %s", machineType)
		}

		memMB, err := strconv.ParseInt(strRAMAmount, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse the custom amount of RAM for %s", machineType)
		}
		memGB := float64(memMB) / 1024.0

		maxMemRatio := 8.0
		if instanceType == "N1" {
			maxMemRatio = 6.5
		}

		extendedMemGB := 0.0
		if extended {
			maxMemGB := float64(cores) * maxMemRatio
			extendedMemGB = math.Max(memGB-maxMemGB, 0)
			if extendedMemGB > 0.0 {
				memGB = maxMemGB
			}
		}

		costComponents := make([]*schema.CostComponent, 0)

		costComponents = append(costComponents, &schema.CostComponent{
			Name:                fmt.Sprintf("Custom instance CPU (Linux/UNIX, %s, %s %d vCPUs)", purchaseOptionLabel(purchaseOption), instanceType, cores),
			Unit:                "hours",
			UnitMultiplier:      decimal.NewFromInt(cores),
			MonthlyQuantity:     decimalPtr(qty.Mul(decimal.NewFromInt(cores))),
			MonthlyDiscountPerc: sustainedUseDiscount,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Compute Engine"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: regexPtr(fmt.Sprintf("^%s%s Instance Core", purchaseOptionPrefix, instanceTypePrefix))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr(fixPurchaseOption),
			},
		})

		costComponents = append(costComponents, &schema.CostComponent{
			Name:                fmt.Sprintf("Custom Instance RAM (Linux/UNIX, %s, %s %s GB)", purchaseOptionLabel(purchaseOption), instanceType, strconv.FormatFloat(memGB, 'f', -1, 64)),
			Unit:                "hours",
			UnitMultiplier:      decimal.NewFromFloat(memGB),
			MonthlyQuantity:     decimalPtr(qty.Mul(decimal.NewFromFloat(memGB))),
			MonthlyDiscountPerc: sustainedUseDiscount,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr(region),
				Service:       strPtr("Compute Engine"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", ValueRegex: regexPtr(fmt.Sprintf("^%s%s Instance Ram", purchaseOptionPrefix, instanceTypePrefix))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr(fixPurchaseOption),
			},
		})

		if extendedMemGB > 0.0 {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:                fmt.Sprintf("Custom Instance Extended RAM (Linux/UNIX, %s, %s %s GB)", purchaseOptionLabel(purchaseOption), instanceType, strconv.FormatFloat(extendedMemGB, 'f', -1, 64)),
				Unit:                "hours",
				UnitMultiplier:      decimal.NewFromFloat(extendedMemGB),
				MonthlyQuantity:     decimalPtr(qty.Mul(decimal.NewFromFloat(extendedMemGB))),
				MonthlyDiscountPerc: sustainedUseDiscount,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("gcp"),
					Region:        strPtr(region),
					Service:       strPtr("Compute Engine"),
					ProductFamily: strPtr("Compute"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "description", ValueRegex: regexPtr(fmt.Sprintf("^%s%s Extended Instance Ram", purchaseOptionPrefix, instanceTypePrefix))},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr(fixPurchaseOption),
				},
			})
		}

		return costComponents, nil
	}

}

func getSustainedUseDiscount(hours float64, rates sudRates) float64 {
	if hours == 0 {
		return 0
	}

	totalHoursInMonth, _ := schema.HourToMonthUnitMultiplier.Float64()

	// Keep track of how many hours we have remaining after each threshold is applied
	remainingHours := hours

	// Keep track of the total hours that are charged for
	ratedHours := 0.0

	index := 0
	for remainingHours > 0 && index < len(rates.rates) {
		// Calculate the percentage of the month that is covered by the current threshold
		lastThreshold := 0.0
		if index > 0 {
			lastThreshold = rates.thresholds[index-1]
		}

		thresholdHours := (rates.thresholds[index] - lastThreshold) * totalHoursInMonth

		// If the remaining hours are less than the threshold, add them and then we are done
		if remainingHours <= thresholdHours {
			ratedHours += remainingHours * rates.rates[index]
			break
		}

		// Otherwise, add the discount for the current threshold and continue
		ratedHours += thresholdHours * rates.rates[index]
		remainingHours -= thresholdHours

		index++
	}

	// Return the average discount over the hours the instance is running
	avgDiscount := 1 - (ratedHours / hours)
	// Round so the calculations match up with Google's 20% discount
	rounded, _ := decimal.NewFromFloat(avgDiscount).Round(2).Float64()
	return rounded
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
	case "pd-extreme":
		diskTypeDesc = "/^Extreme PD Capacity/"
		diskTypeLabel = "Extreme provisioned storage (pd-extreme)"
	case "hyperdisk-extreme":
		diskTypeDesc = "/^Hyperdisk Extreme Capacity( in .*)?$/"
		diskTypeLabel = "Hyperdisk provisioned storage (hyperdisk-extreme)"
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

func computeDiskIOPSCostComponent(region string, diskType string, diskSize float64, instanceCount int64, iops int64) *schema.CostComponent {
	var iopsTypeDesc string

	switch diskType {
	case "pd-extreme":
		iopsTypeDesc = "/^Extreme PD IOPS/"
	case "hyperdisk-extreme":
		iopsTypeDesc = "/^Hyperdisk Extreme IOPS( in .*)?$/"
	}

	return &schema.CostComponent{
		Name:            "Provisioned IOPS",
		Unit:            "IOPS",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(iops)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: strPtr(iopsTypeDesc)},
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
	// From strings in format 'nvidia-tesla-a100' create:
	// - name: 'NVIDIA Tesla A100'
	// - descPrefix: 'Nvidia Tesla A100 GPU'
	parts := strings.Split(guestAcceleratorType, "-")
	if len(parts) < 2 {
		logging.Logger.Debug().Msgf("skipping cost component because guest_accelerator.type '%s' is not supported", guestAcceleratorType)
		return nil
	}

	caser := cases.Title(language.English)
	rest := caser.String(strings.Join(parts[1:], " "))

	name := fmt.Sprintf("%s %s", strings.ToUpper(parts[0]), rest)
	descPrefix := fmt.Sprintf("%s %s GPU", caser.String(parts[0]), rest)

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
		UsageBased: true,
	}
}
