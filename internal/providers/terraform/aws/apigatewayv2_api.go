package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"strings"
)

func GetAPIGatewayv2ApiRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_apigatewayv2_api",
		RFunc: NewAPIGatewayv2Api,
	}
}

func NewAPIGatewayv2Api(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent

	protocolType := d.Get("protocol_type").String()

	if strings.ToLower(protocolType) == "websocket" {
		costComponents = websocketAPICostComponent(d, u)
	}

	if strings.ToLower(protocolType) == "http" {
		costComponents = httpAPICostComponent(d, u)
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func httpAPICostComponent(d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {
	region := d.Get("region").String()
	var monthlyRequests *decimal.Decimal
	requestSize := decimal.NewFromInt(512)

	billableRequestSize := decimal.NewFromInt(512)

	httpAPITiers := []int{300000000}

	costComponents := []*schema.CostComponent{}

	if u != nil && u.Get("monthly_requests").Exists() {
		monthlyRequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_requests").Int()))

		if u.Get("request_size_kb").Exists() {
			requestSize = decimal.NewFromInt(u.Get("request_size_kb").Int())
		}

		if requestSize.GreaterThan(billableRequestSize) {
			monthlyRequests = calculateBillableRequests(&requestSize, &billableRequestSize, monthlyRequests)
		}

		apiRequestQuantities := usage.CalculateTierBuckets(*monthlyRequests, httpAPITiers)

		costComponents = append(costComponents, httpCostComponent(region, "Requests (first 300M)", "0", &apiRequestQuantities[0]))

		if apiRequestQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, httpCostComponent(region, "Requests (over 300M)", "300000000", &apiRequestQuantities[1]))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, httpCostComponent(region, "Requests (first 300M)", "0", unknown))
	}

	return costComponents
}

func websocketAPICostComponent(d *schema.ResourceData, u *schema.UsageData) []*schema.CostComponent {
	region := d.Get("region").String()
	var monthlyMessages *decimal.Decimal
	var monthlyConnectionMinutes *decimal.Decimal

	messageSize := decimal.NewFromInt(32)

	billableRequestSize := decimal.NewFromInt(32)

	// Websocket request tiers
	websocketAPITiers := []int{1000000000}

	costComponents := []*schema.CostComponent{}

	if u != nil && u.Get("monthly_messages").Exists() {
		monthlyMessages = decimalPtr(decimal.NewFromInt(u.Get("monthly_messages").Int()))

		if u.Get("message_size_kb").Exists() {
			messageSize = decimal.NewFromInt(u.Get("message_size_kb").Int())
		}

		if messageSize.GreaterThan(billableRequestSize) {
			monthlyMessages = calculateBillableRequests(&messageSize, &billableRequestSize, monthlyMessages)
		}

		apiRequestQuantities := usage.CalculateTierBuckets(*monthlyMessages, websocketAPITiers)

		costComponents = append(costComponents, websocketCostComponent(region, "messages", "ApiGatewayMessage", "Messages (first 1B)", "0", &apiRequestQuantities[0]))

		if apiRequestQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, websocketCostComponent(region, "messages", "ApiGatewayMessage", "Messages (over 1B)", "1000000000", &apiRequestQuantities[1]))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, websocketCostComponent(region, "messages", "ApiGatewayMessage", "Messages (first 1B)", "0", unknown))
	}

	if u != nil && u.Get("monthly_connection_mins").Exists() {
		monthlyConnectionMinutes = decimalPtr(decimal.NewFromInt(u.Get("monthly_connection_mins").Int()))
	}
	costComponents = append(costComponents, websocketCostComponent(region, "minutes", "ApiGatewayMinute", "Connection duration", "0", monthlyConnectionMinutes))

	return costComponents
}

func calculateBillableRequests(requestSize *decimal.Decimal, billableRequestSize *decimal.Decimal, requests *decimal.Decimal) *decimal.Decimal {
	return decimalPtr(requests.Mul(requestSize.Div(*billableRequestSize).Ceil()))
}

func httpCostComponent(region string, displayName string, usageTier string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "1M requests",
		UnitMultiplier:  1000000,
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonApiGateway"),
			ProductFamily: strPtr("API Calls"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayHttpRequest/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}

func websocketCostComponent(region string, unit string, usageType string, displayName string, usageTier string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "1M " + unit,
		UnitMultiplier:  1000000,
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonApiGateway"),
			ProductFamily: strPtr("WebSocket"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/" + usageType + "/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
