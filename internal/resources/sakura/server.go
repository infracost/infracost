package sakura

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// serverPlanKey identifies a server plan by CPU core count and memory (GiB).
type serverPlanKey struct {
	Core   int64
	Memory int64
}

// serverMonthlyPriceJPY maps (core, memory_GiB) to monthly price in JPY (tax-excluded).
// Prices sourced from https://cloud.sakura.ad.jp/products/server/ (石狩第1ゾーン).
var serverMonthlyPriceJPY = map[serverPlanKey]float64{
	{1, 1}:   1400,
	{1, 2}:   1950,
	{2, 2}:   3000,
	{2, 4}:   4200,
	{4, 4}:   6000,
	{4, 8}:   8400,
	{8, 8}:   12000,
	{8, 16}:  16800,
	{12, 12}: 18000,
	{16, 16}: 24000,
	{20, 32}: 37200,
}

// dedicatedIntelMonthlyPriceJPY maps (core, memory_GiB) to monthly price in JPY (tax-excluded).
// Prices sourced from https://cloud.sakura.ad.jp/products/server/dedicated-core/ (Intel Xeon).
var dedicatedIntelMonthlyPriceJPY = map[serverPlanKey]float64{
	{2, 4}:    9200,
	{4, 8}:    18400,
	{4, 16}:   23200,
	{6, 32}:   39600,
	{8, 16}:   36800,
	{8, 32}:   46400,
	{10, 24}:  48400,
	{10, 32}:  53200,
	{10, 48}:  62800,
	{10, 96}:  91600,
	{24, 192}: 200000,
}

// dedicatedAMDMonthlyPriceJPY maps (core, memory_GiB) to monthly price in JPY (tax-excluded).
// Prices sourced from https://cloud.sakura.ad.jp/products/server/dedicated-core/amd/ (AMD EPYC).
var dedicatedAMDMonthlyPriceJPY = map[serverPlanKey]float64{
	{32, 120}:   125000,
	{64, 240}:   250000,
	{128, 480}:  500000,
	{192, 1024}: 1100000,
}

// gpuMonthlyPriceJPY maps GPU model name (lowercase) to monthly price in JPY (tax-excluded).
// Prices sourced from https://cloud.sakura.ad.jp/products/server/gpu/ (高火力VRT, 石狩第1ゾーン).
var gpuMonthlyPriceJPY = map[string]float64{
	"v100": 210000,
	"h100": 350000,
}

// Server represents a sakura_server Terraform resource.
type Server struct {
	Address    string
	Zone       string
	Core       int64
	Memory     int64
	Commitment string
	CPUModel   string
	GPUCount   int64
	GPUModel   string
}

func (r *Server) CoreType() string {
	return "Server"
}

func (r *Server) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *Server) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *Server) BuildResource() *schema.Resource {
	core := r.Core
	if core == 0 {
		core = 1
	}
	memory := r.Memory
	if memory == 0 {
		memory = 1
	}

	var cc *schema.CostComponent

	switch {
	case r.GPUCount > 0:
		cc = r.gpuCostComponent()
	case strings.ToLower(r.Commitment) == "dedicated":
		cc = r.dedicatedCostComponent(core, memory)
	default:
		cc = r.standardCostComponent(core, memory)
	}

	if cc == nil {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{cc},
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *Server) standardCostComponent(core, memory int64) *schema.CostComponent {
	key := serverPlanKey{Core: core, Memory: memory}
	price, ok := serverMonthlyPriceJPY[key]
	if !ok {
		return nil
	}
	cc := &schema.CostComponent{
		Name:            fmt.Sprintf("Server (%d core, %dGB)", core, memory),
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
	}
	cc.SetCustomPrice(decimalPtr(decimal.NewFromFloat(price)))
	return cc
}

func (r *Server) dedicatedCostComponent(core, memory int64) *schema.CostComponent {
	key := serverPlanKey{Core: core, Memory: memory}

	// AMD EPYC models have "epyc" or "amd" in the cpu_model string.
	isAMD := strings.Contains(strings.ToLower(r.CPUModel), "epyc") ||
		strings.Contains(strings.ToLower(r.CPUModel), "amd")

	var price float64
	var ok bool
	var label string

	if isAMD {
		price, ok = dedicatedAMDMonthlyPriceJPY[key]
		label = fmt.Sprintf("Server dedicated AMD (%d core, %dGB)", core, memory)
	} else {
		price, ok = dedicatedIntelMonthlyPriceJPY[key]
		label = fmt.Sprintf("Server dedicated Intel (%d core, %dGB)", core, memory)
	}

	if !ok {
		return nil
	}

	cc := &schema.CostComponent{
		Name:            label,
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
	}
	cc.SetCustomPrice(decimalPtr(decimal.NewFromFloat(price)))
	return cc
}

func (r *Server) gpuCostComponent() *schema.CostComponent {
	model := strings.ToLower(r.GPUModel)
	price, ok := gpuMonthlyPriceJPY[model]
	if !ok {
		return nil
	}
	cc := &schema.CostComponent{
		Name:            fmt.Sprintf("Server GPU %s (×%d)", strings.ToUpper(r.GPUModel), r.GPUCount),
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(r.GPUCount)),
	}
	cc.SetCustomPrice(decimalPtr(decimal.NewFromFloat(price)))
	return cc
}
