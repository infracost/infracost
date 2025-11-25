package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

func GetAzureRMSearchServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_search_service",
		RFunc: NewAzureRMSearchService,
		ReferenceAttributes: []string{
			"resource_group_name",
		},
	}
}

func NewAzureRMSearchService(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region
	costComponents := []*schema.CostComponent{}

	sku := strings.ToLower(d.Get("sku").String())
	if sku == "free" {
		return &schema.Resource{
			Name:      d.Address,
			NoPrice:   true,
			IsSkipped: true,
		}
	}

	if strings.HasPrefix(sku, "standard") {
		sku = sku[:len(sku)-1] + " s" + sku[len(sku)-1:]
	}
	if strings.HasPrefix(sku, "storage") {
		sku = strings.ReplaceAll(sku, "_", " ")
	}

	partitionCount := decimal.NewFromInt(1)
	replicaCount := decimal.NewFromInt(1)

	if d.Get("partition_count").Type != gjson.Null {
		partitionCount = decimal.NewFromInt(d.Get("partition_count").Int())
	}
	if d.Get("replica_count").Type != gjson.Null {
		replicaCount = decimal.NewFromInt(d.Get("replica_count").Int())
	}
	units := decimalPtr(partitionCount.Mul(replicaCount))

	var skuName string
	skuElems := strings.SplitSeq(sku, " ")
	for v := range skuElems {
		skuName += cases.Title(language.English).String(v) + " "
	}
	unitName := "unit"
	if units.GreaterThan(decimal.NewFromInt(1)) {
		unitName += "s"
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:           fmt.Sprintf("Search usage (%s, %s %s)", skuName[:len(skuName)-1], units.String(), unitName),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: units,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Cognitive Search"),
			ProductFamily: strPtr("Web"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", sku))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	})

	var images *decimal.Decimal
	if u != nil && u.Get("monthly_images_extracted").Type != gjson.Null {
		images = decimalPtr(decimal.NewFromInt(u.Get("monthly_images_extracted").Int()))
		tierLimits := []int{1_000_000, 4_000_000}
		tiers := usage.CalculateTierBuckets(*images, tierLimits)

		type dataTier struct {
			name       string
			startUsage string
		}

		data := []dataTier{
			{name: "Image extraction (first 1M)", startUsage: "0"},
			{name: "Image extraction (next 4M)", startUsage: "1000"},
			{name: "Image extraction (over 5M)", startUsage: "5000"},
		}
		for i, d := range data {
			if tiers[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, searchServiceCostComponent(
					region,
					d.name,
					d.startUsage,
					decimalPtr(tiers[i].Div(decimal.NewFromInt(1000)))))
			}
		}
	} else {
		costComponents = append(costComponents, searchServiceCostComponent(
			region,
			"Image extraction (first 1M)",
			"0",
			images))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func searchServiceCostComponent(region, name, startUsage string, qty *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "1000 images",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: qty,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Azure Cognitive Search"),
			ProductFamily: strPtr("Web"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr("Document Cracking")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
