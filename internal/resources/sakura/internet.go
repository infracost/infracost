package sakura

import (
	"fmt"
	"sort"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// internetBandwidthPricingJPY maps bandwidth (Mbps) to monthly price in JPY (tax-excluded).
// Prices sourced from https://cloud.sakura.ad.jp/products/router-switch/ (石狩・東京第1ゾーン).
var internetBandwidthPricingJPY = []bandwidthPriceTier{
	{100, 2000},
	{250, 24000},
	{500, 48000},
	{1000, 98000},
	{1500, 148000},
	{2000, 198000},
	{2500, 248000},
	{3000, 298000},
	{3500, 348000},
	{4000, 398000},
	{4500, 448000},
	{5000, 498000},
}

// switchMonthlyPriceJPY is the monthly price for a switch (全ゾーン共通, tax-excluded).
const switchMonthlyPriceJPY = 2000

type bandwidthPriceTier struct {
	BandwidthMbps int64
	Price         float64
}

func lookupBandwidthPrice(mbps int64) (float64, int64) {
	tiers := internetBandwidthPricingJPY
	sort.Slice(tiers, func(i, j int) bool { return tiers[i].BandwidthMbps < tiers[j].BandwidthMbps })
	for _, t := range tiers {
		if mbps <= t.BandwidthMbps {
			return t.Price, t.BandwidthMbps
		}
	}
	last := tiers[len(tiers)-1]
	return last.Price, last.BandwidthMbps
}

// Internet represents a sakura_internet (ルータ+スイッチ) Terraform resource.
type Internet struct {
	Address   string
	Zone      string
	BandWidth int64
}

func (r *Internet) CoreType() string {
	return "Internet"
}

func (r *Internet) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *Internet) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Internet) BuildResource() *schema.Resource {
	bw := r.BandWidth
	if bw == 0 {
		bw = 100
	}

	monthlyPrice, matchedBW := lookupBandwidthPrice(bw)

	routerCC := &schema.CostComponent{
		Name:            fmt.Sprintf("ルータ+スイッチ (%dMbps)", matchedBW),
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
	}
	routerCC.SetCustomPrice(decimalPtr(decimal.NewFromFloat(monthlyPrice)))

	switchCC := &schema.CostComponent{
		Name:            "スイッチ",
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
	}
	switchCC.SetCustomPrice(decimalPtr(decimal.NewFromFloat(switchMonthlyPriceJPY)))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{routerCC, switchCC},
		UsageSchema:    r.UsageSchema(),
	}
}
