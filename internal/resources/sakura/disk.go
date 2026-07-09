package sakura

import (
	"fmt"
	"sort"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// ssdPricingJPY maps SSD disk size (GiB) to monthly price in JPY (tax-excluded).
// Prices sourced from https://cloud.sakura.ad.jp/products/disk/ (石狩第1ゾーン).
var ssdPricingJPY = []diskPriceTier{
	{20, 400},
	{40, 1400},
	{100, 3500},
	{250, 8750},
	{500, 17500},
	{1024, 30000},
	{2048, 60000},
	{4096, 120000},
	{8192, 240000},
	{12288, 360000},
	{16384, 480000},
}

// hddPricingJPY maps HDD disk size (GiB) to monthly price in JPY (tax-excluded).
// Prices sourced from https://cloud.sakura.ad.jp/products/disk/ (石狩第3/東京第1ゾーン).
var hddPricingJPY = []diskPriceTier{
	{40, 800},
	{2048, 30000},
	{4096, 60000},
	{8192, 120000},
	{12288, 180000},
}

type diskPriceTier struct {
	SizeGiB int64
	Price   float64
}

func lookupDiskPrice(tiers []diskPriceTier, sizeGiB int64) (float64, int64) {
	sort.Slice(tiers, func(i, j int) bool { return tiers[i].SizeGiB < tiers[j].SizeGiB })
	for _, t := range tiers {
		if sizeGiB <= t.SizeGiB {
			return t.Price, t.SizeGiB
		}
	}
	last := tiers[len(tiers)-1]
	return last.Price, last.SizeGiB
}

// Disk represents a sakura_disk Terraform resource.
type Disk struct {
	Address string
	Zone    string
	// Plan is "ssd" or "hdd" (default: "ssd")
	Plan string
	// Size is disk size in GiB (default: 20)
	Size int64
}

func (r *Disk) CoreType() string {
	return "Disk"
}

func (r *Disk) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *Disk) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Disk) BuildResource() *schema.Resource {
	plan := r.Plan
	if plan == "" {
		plan = "ssd"
	}
	size := r.Size
	if size == 0 {
		size = 20
	}

	var tiers []diskPriceTier
	switch plan {
	case "hdd":
		tiers = hddPricingJPY
	default:
		tiers = ssdPricingJPY
	}

	monthlyPrice, matchedSize := lookupDiskPrice(tiers, size)

	cc := &schema.CostComponent{
		Name:            fmt.Sprintf("Disk (%s, %dGiB)", plan, matchedSize),
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
	}
	cc.SetCustomPrice(decimalPtr(decimal.NewFromFloat(monthlyPrice)))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{cc},
		UsageSchema:    r.UsageSchema(),
	}
}
