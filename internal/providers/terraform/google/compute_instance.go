package google

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"

	"github.com/shopspring/decimal"
)

func GetComputeInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_instance",
		RFunc: NewComputeInstance,
		Notes: []string{
			"Sustained use discounts are applied to monthly costs, but not to hourly costs.",
			"Costs associated with non-standard Linux images, such as Windows and RHEL are not supported.",
			"Custom machine types are not supported.",
			"Sole-tenant VMs are not supported.",
		},
	}
}

func NewComputeInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	machineType := d.Get("machine_type").String()
	if strings.HasPrefix(machineType, "custom-") {
		return nil
	}

	region := d.Get("region").String()

	zone := d.Get("zone").String()
	if zone != "" {
		region = zoneToRegion(zone)
	}

	purchaseOption := "on_demand"
	if d.Get("scheduling.0.preemptible").Bool() {
		purchaseOption = "preemptible"
	}

	costComponents := []*schema.CostComponent{computeCostComponent(region, machineType, purchaseOption)}

	if d.Get("boot_disk.0.initialize_params.0").Exists() {
		costComponents = append(costComponents, bootDisk(region, d.Get("boot_disk.0.initialize_params.0")))
	}

	if len(d.Get("scratch_disk").Array()) > 0 {
		count := len(d.Get("scratch_disk").Array())
		costComponents = append(costComponents, scratchDisk(region, purchaseOption, count))
	}

	for _, guestAccel := range d.Get("guest_accelerator").Array() {
		costComponents = append(costComponents, guestAccelerator(region, purchaseOption, guestAccel))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func computeCostComponent(region, machineType string, purchaseOption string) *schema.CostComponent {
	sustainedUseDiscount := 0.0
	if purchaseOption == "on_demand" {
		switch strings.Split(machineType, "-")[0] {
		case "c2", "n2", "n2d":
			sustainedUseDiscount = 0.2
		case "n1", "f1", "g1", "m1":
			sustainedUseDiscount = 0.3
		}
	}

	return &schema.CostComponent{
		Name:                fmt.Sprintf("Instance usage (Linux/UNIX, %s, %s)", purchaseOptionLabel(purchaseOption), machineType),
		Unit:                "hours",
		UnitMultiplier:      1,
		HourlyQuantity:      decimalPtr(decimal.NewFromInt(1)),
		MonthlyDiscountPerc: sustainedUseDiscount,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Compute Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "machineType", Value: strPtr(machineType)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr(purchaseOption),
		},
	}
}

func bootDisk(region string, initializeParams gjson.Result) *schema.CostComponent {
	size := decimalPtr(decimal.NewFromInt(int64(defaultVolumeSize)))
	if initializeParams.Get("size").Exists() {
		size = decimalPtr(decimal.NewFromFloat(initializeParams.Get("size").Float()))
	}

	diskType := initializeParams.Get("type").String()

	return computeDisk(region, diskType, size)
}

func scratchDisk(region string, purchaseOption string, count int) *schema.CostComponent {
	descRegex := "/^SSD backed Local Storage( in .*)?$/"
	if purchaseOption == "preemptible" {
		descRegex = "/^SSD backed Local Storage attached to Preemptible VMs/"
	}

	return &schema.CostComponent{
		Name:            "Local SSD provisioned storage",
		Unit:            "GiB",
		UnitMultiplier:  1,
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(int64(375 * count))), // local SSDs are always 375 GiB
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

func guestAccelerator(region string, purchaseOption string, guestAccel gjson.Result) *schema.CostComponent {
	model := guestAccel.Get("type").String()

	var (
		name       string
		descPrefix string
	)

	switch model {
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
	default:
		return nil
	}

	descRegex := fmt.Sprintf("/^%s running/", descPrefix)
	if purchaseOption == "preemptible" {
		descRegex = fmt.Sprintf("/^%s attached to preemptible VMs running/", descPrefix)
	}

	count := decimal.NewFromInt(guestAccel.Get("count").Int())

	sustainedUseDiscount := 0.0
	if purchaseOption == "on_demand" {
		sustainedUseDiscount = 0.3
	}

	return &schema.CostComponent{
		Name:                fmt.Sprintf("%s (%s)", name, purchaseOptionLabel(purchaseOption)),
		Unit:                "hours",
		UnitMultiplier:      1,
		HourlyQuantity:      decimalPtr(count),
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

func purchaseOptionLabel(purchaseOption string) string {
	return map[string]string{
		"on_demand":   "on-demand",
		"preemptible": "preemptible",
	}[purchaseOption]
}
