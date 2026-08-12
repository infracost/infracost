package sakura

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// apprunSharedPricingTier holds a pricing tier for AppRun shared plan.
// CPU is in cores (e.g. 0.5, 1.0, 2.0), Memory is in GiB.
// Prices sourced from https://cloud.sakura.ad.jp/products/apprun-shared/ (税抜).
type apprunSharedPricingTier struct {
	CPU        float64
	MemoryGiB  float64
	HourlyJPY  float64
}

// apprunSharedTiers is sorted ascending by (CPU, Memory) so the first matching tier is used.
var apprunSharedTiers = []apprunSharedPricingTier{
	{0.5, 1.0, 4.545},
	{1.0, 1.0, 7.272},
	{1.0, 2.0, 8.181},
	{2.0, 2.0, 15.454},
	{2.0, 4.0, 17.272},
}

// AppComponent holds per-component resource requests.
type AppComponent struct {
	MaxCPU    string
	MaxMemory string
}

// ApprunShared represents a sakura_apprun_shared Terraform resource.
type ApprunShared struct {
	Address    string
	MinScale   int64
	MaxScale   int64
	Components []AppComponent
}

func (r *ApprunShared) CoreType() string {
	return "ApprunShared"
}

func (r *ApprunShared) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ApprunShared) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ApprunShared) BuildResource() *schema.Resource {
	maxScale := r.MaxScale
	if maxScale == 0 {
		maxScale = 1
	}

	totalCPU, totalMemGiB := r.totalResources()
	tier := lookupApprunTier(totalCPU, totalMemGiB)

	// Upper bound: all instances at max_scale running for a full month (730h).
	monthlyUpperBound := decimal.NewFromFloat(tier.HourlyJPY).
		Mul(decimal.NewFromInt(730)).
		Mul(decimal.NewFromInt(maxScale))

	cc := &schema.CostComponent{
		Name: fmt.Sprintf(
			"AppRun shared (%.1f core, %.1fGiB × max %d instances, upper bound)",
			tier.CPU, tier.MemoryGiB, maxScale,
		),
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		UsageBased:      true,
	}
	cc.SetCustomPrice(decimalPtr(monthlyUpperBound))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{cc},
		UsageSchema:    r.UsageSchema(),
	}
}

// totalResources sums max_cpu and max_memory across all components.
func (r *ApprunShared) totalResources() (float64, float64) {
	var totalCPU, totalMemGiB float64
	for _, c := range r.Components {
		totalCPU += parseCPU(c.MaxCPU)
		totalMemGiB += parseMemoryGiB(c.MaxMemory)
	}
	if totalCPU == 0 {
		totalCPU = 0.5
	}
	if totalMemGiB == 0 {
		totalMemGiB = 1.0
	}
	return totalCPU, totalMemGiB
}

// lookupApprunTier returns the first tier whose CPU >= totalCPU and Memory >= totalMemGiB.
// Falls back to the largest tier if none matches.
func lookupApprunTier(totalCPU, totalMemGiB float64) apprunSharedPricingTier {
	for _, t := range apprunSharedTiers {
		if t.CPU >= totalCPU && t.MemoryGiB >= totalMemGiB {
			return t
		}
	}
	return apprunSharedTiers[len(apprunSharedTiers)-1]
}

// parseCPU converts strings like "0.5", "1", "2" to float64 core count.
func parseCPU(s string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}

// parseMemoryGiB converts Kubernetes memory strings ("256Mi", "1Gi", "2Gi") to GiB.
func parseMemoryGiB(s string) float64 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "Gi") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "Gi"), 64)
		if err != nil {
			return 0
		}
		return v
	}
	if strings.HasSuffix(s, "Mi") {
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "Mi"), 64)
		if err != nil {
			return 0
		}
		return v / 1024
	}
	// bare number → assume GiB
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
