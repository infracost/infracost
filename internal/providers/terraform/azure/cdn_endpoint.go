package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

func GetAzureRMCDNEndpointRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_cdn_endpoint",
		RFunc: NewAzureRMCDNEndpoint,
		ReferenceAttributes: []string{
			"profile_name",
		},
	}
}

func NewAzureRMCDNEndpoint(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := regionToCDNZone(d.Region)

	var costComponents []*schema.CostComponent

	sku := ""
	var profile *schema.ResourceData
	if len(d.References("profile_name")) > 0 {
		profile = d.References("profile_name")[0]
		sku = profile.Get("sku").String()
	}

	if len(strings.Split(sku, "_")) != 2 || strings.ToLower(sku) == "standard_chinacdn" {
		logging.Logger.Warn().Msgf("Unrecognized/unsupported CDN sku format for resource %s: %s", d.Address, sku)
		return nil
	}

	costComponents = append(costComponents, cdnOutboundDataCostComponents(region, sku, u)...)

	if strings.ToLower(sku) == "standard_microsoft" {
		numberOfRules := 0
		if d.Get("global_delivery_rule").Type != gjson.Null {
			numberOfRules += len(d.Get("global_delivery_rule").Array())
		}
		if d.Get("delivery_rule").Type != gjson.Null {
			numberOfRules += len(d.Get("delivery_rule").Array())
		}

		if numberOfRules > 5 {
			numberOfRules -= 5

			costComponents = append(costComponents, cdnCostComponent(
				"Rules engine rules (over 5)",
				"rules",
				region,
				"Azure CDN from Microsoft",
				"Standard",
				"Rule",
				"5",
				decimalPtr(decimal.NewFromInt(int64(numberOfRules))),
			))
		}

		if numberOfRules > 0 {
			var rulesRequests *decimal.Decimal
			if u != nil && u.Get("monthly_rules_engine_requests").Type != gjson.Null {
				rulesRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_rules_engine_requests").Int() / 1000000))
			}
			costComponents = append(costComponents, cdnCostComponent(
				"Rules engine requests",
				"1M requests",
				region,
				"Azure CDN from Microsoft",
				"Standard",
				"Requests",
				"0",
				rulesRequests,
			))
		}
	}

	if strings.ToLower(sku) == "standard_akamai" || strings.ToLower(sku) == "standard_verizon" {
		if d.Get("optimization_type").Type != gjson.Null {
			if strings.ToLower(d.Get("optimization_type").String()) == "dynamicsiteacceleration" {
				costComponents = append(costComponents, cdnAccelerationDataTransfersCostComponents(region, sku, d, u)...)
			}
		}
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func cdnOutboundDataCostComponents(region, sku string, u *schema.UsageData) []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	type dataTier struct {
		name       string
		startUsage string
	}

	var name, productName, skuName, meterName string
	if s := strings.Split(sku, "_"); len(s) == 2 {
		productName = fmt.Sprintf("Azure CDN from %s", s[1])
		skuName = s[0]
		if strings.ToLower(s[1]) == "verizon" {
			name = fmt.Sprintf("Outbound data transfer (%s, ", s[0]+" "+s[1])
		} else {
			name = fmt.Sprintf("Outbound data transfer (%s, ", s[1])
		}
	}

	data := []dataTier{
		{name: fmt.Sprintf("%s%s", name, "first 10TB)"), startUsage: "0"},
		{name: fmt.Sprintf("%s%s", name, "next 40TB)"), startUsage: "10000"},
		{name: fmt.Sprintf("%s%s", name, "next 100TB)"), startUsage: "50000"},
		{name: fmt.Sprintf("%s%s", name, "next 350TB)"), startUsage: "150000"},
		{name: fmt.Sprintf("%s%s", name, "next 500TB)"), startUsage: "500000"},
		{name: fmt.Sprintf("%s%s", name, "next 4000TB)"), startUsage: "1000000"},
		{name: fmt.Sprintf("%s%s", name, "over 5000TB)"), startUsage: "5000000"},
	}

	meterName = fmt.Sprintf("%s Data Transfer", skuName)

	var monthlyOutboundGb *decimal.Decimal
	if u != nil && u.Get("monthly_outbound_gb").Type != gjson.Null {
		monthlyOutboundGb = decimalPtr(decimal.NewFromInt(u.Get("monthly_outbound_gb").Int()))
		tierLimits := []int{10000, 40000, 100000, 350000, 500000, 4000000}
		tiers := usage.CalculateTierBuckets(*monthlyOutboundGb, tierLimits)

		for i, d := range data {
			if tiers[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, cdnCostComponent(
					d.name,
					"GB",
					region,
					productName,
					skuName,
					meterName,
					d.startUsage,
					decimalPtr(tiers[i])))
			}
		}
	} else {
		costComponents = append(costComponents, cdnCostComponent(
			data[0].name,
			"GB",
			region,
			productName,
			skuName,
			meterName,
			data[0].startUsage,
			nil))
	}

	return costComponents
}

func cdnAccelerationDataTransfersCostComponents(region, sku string, d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	type dataTier struct {
		name       string
		startUsage string
	}

	name := "Acceleration outbound data transfer "

	data := []dataTier{
		{name: fmt.Sprintf("%s%s", name, "(first 50TB)"), startUsage: "0"},
		{name: fmt.Sprintf("%s%s", name, "(next 100TB)"), startUsage: "50000"},
		{name: fmt.Sprintf("%s%s", name, "(next 350TB)"), startUsage: "150000"},
		{name: fmt.Sprintf("%s%s", name, "(next 500TB)"), startUsage: "500000"},
		{name: fmt.Sprintf("%s%s", name, "(over 1000TB)"), startUsage: "1000000"},
	}

	var productName, skuName, meterName string
	if s := strings.Split(sku, "_"); len(s) == 2 {
		productName = fmt.Sprintf("Azure CDN from %s", s[1])
		skuName = s[0]
	}
	meterName = "Standard Acceleration Data Transfer"

	var monthlyOutboundGb *decimal.Decimal
	if u != nil && u.Get("monthly_outbound_gb").Type != gjson.Null {
		monthlyOutboundGb = decimalPtr(decimal.NewFromInt(u.Get("monthly_outbound_gb").Int()))
		tierLimits := []int{50000, 100000, 350000, 500000, 1000000}
		tiers := usage.CalculateTierBuckets(*monthlyOutboundGb, tierLimits)

		for i, d := range data {
			if tiers[i].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, cdnCostComponent(
					d.name,
					"GB",
					region,
					productName,
					skuName,
					meterName,
					d.startUsage,
					decimalPtr(tiers[i])))
			}
		}
	} else {
		costComponents = append(costComponents, cdnCostComponent(
			data[0].name,
			"GB",
			region,
			productName,
			skuName,
			meterName,
			data[0].startUsage,
			nil))
	}

	return costComponents
}

func cdnCostComponent(name, unit, region, productName, skuName, meterName, startUsage string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            unit,
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("Content Delivery Network"),
			ProductFamily: strPtr("Networking"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", productName))},
				{Key: "skuName", ValueRegex: regexPtr(fmt.Sprintf("^%s$", skuName))},
				{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s$", meterName))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}
