package sakura

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// privateHostMonthlyPriceJPY maps class name to monthly price in JPY (tax-excluded).
// Prices sourced from https://cloud.sakura.ad.jp/products/server/dedicated-host/.
var privateHostMonthlyPriceJPY = map[string]float64{
	"dynamic":           200000,
	"windows":           230000,
	"dedicated_storage": 1100000,
}

// PrivateHost represents a sakura_private_host Terraform resource.
type PrivateHost struct {
	Address            string
	Zone               string
	Class              string
	DedicatedStorageID string
}

func (r *PrivateHost) CoreType() string {
	return "PrivateHost"
}

func (r *PrivateHost) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *PrivateHost) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *PrivateHost) BuildResource() *schema.Resource {
	class := strings.ToLower(r.Class)
	if class == "" {
		class = "dynamic"
	}

	// dedicated_storage_id が指定されている場合は専有ストレージ用ホスト料金
	key := class
	if r.DedicatedStorageID != "" {
		key = "dedicated_storage"
	}

	price, ok := privateHostMonthlyPriceJPY[key]
	if !ok {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	cc := &schema.CostComponent{
		Name:            fmt.Sprintf("Private host (%s)", key),
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
	}
	cc.SetCustomPrice(decimalPtr(decimal.NewFromFloat(price)))

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{cc},
		UsageSchema:    r.UsageSchema(),
	}
}
