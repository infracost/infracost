package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetApiGatewayv2ApiRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_apigatewayv2_api",
		Notes: []string{
			"WebSocket Connection minutes is not yet supported",
		},
		RFunc: NewApiGatewayv2Api,
	}
}

func NewApiGatewayv2Api(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	var costComponents []*schema.CostComponent

	protocolType := d.Get("protocol_type").String()

	if protocolType == "WEBSOCKET" {
		costComponents = websocketApiCostComponent(d, u)
	}

	if protocolType == "HTTP" {
		costComponents = httpApiCostComponent(d, u)
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func httpApiCostComponent(d *schema.ResourceData, u *schema.ResourceData) []*schema.CostComponent {
	region := d.Get("region").String()
	monthlyRequests := decimal.Zero
	requestSize := decimal.NewFromInt(0)

	billableRequestSize := decimal.NewFromInt(512)

	var apiTierRequests = map[string]decimal.Decimal{
		"apiRequestTierOne": decimal.Zero,
		"apiRequestTierTwo": decimal.Zero,
	}

	// httpApi request tiers
	apiTierOneLimit := decimal.NewFromInt(300000000)
	apiTierTwoLimit := decimal.NewFromInt(300000001)

	if u != nil && u.Get("monthly_requests.0.value").Exists() {
		monthlyRequests = decimal.NewFromInt(u.Get("monthly_requests.0.value").Int())
	}

	if u != nil && u.Get("request_size.0.value").Exists() {
		requestSize = decimal.NewFromInt(u.Get("request_size.0.value").Int())
	}

	if requestSize.GreaterThan(billableRequestSize) {
		monthlyRequests = calculateBillableRequests(requestSize, billableRequestSize, monthlyRequests)
	}

	apiRequestQuantities := calculateApiRequests(monthlyRequests, apiTierRequests, apiTierOneLimit, apiTierTwoLimit)

	apiTierOne := apiRequestQuantities["apiRequestTierOne"]
	apiTierTwo := apiRequestQuantities["apiRequestTierTwo"]

	return []*schema.CostComponent{
		{
			Name:            "Requests (first 300m)",
			Unit:            "requests",
			MonthlyQuantity: &apiTierOne,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonApiGateway"),
				ProductFamily: strPtr("API Calls"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr("HTTP API Requests")},
					{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayHttpRequest/")},
					{Key: "operation", Value: strPtr("ApiGatewayHttpApi")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("0"),
			},
		},
		{
			Name:            "Requests (over 300m)",
			Unit:            "requests",
			MonthlyQuantity: &apiTierTwo,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonApiGateway"),
				ProductFamily: strPtr("API Calls"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr("HTTP API Requests")},
					{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayHttpRequest/")},
					{Key: "operation", Value: strPtr("ApiGatewayHttpApi")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("300000000"),
			},
		},
	}
}

func websocketApiCostComponent(d *schema.ResourceData, u *schema.ResourceData) []*schema.CostComponent {
	region := d.Get("region").String()
	monthlyMessages := decimal.Zero
	messageSize := decimal.NewFromInt(0)

	billableRequestSize := decimal.NewFromInt(32)

	var apiTierRequests = map[string]decimal.Decimal{
		"apiRequestTierOne": decimal.Zero,
		"apiRequestTierTwo": decimal.Zero,
	}

	// Websocket request tiers
	apiTierOneLimt := decimal.NewFromInt(1000000000)
	apiTierTwoLimit := decimal.NewFromInt(1000000001)

	if u != nil && u.Get("monthly_messages.0.value").Exists() {
		monthlyMessages = decimal.NewFromInt(u.Get("monthly_messages.0.value").Int())
	}

	if u != nil && u.Get("average_message_size.0.value").Exists() {
		messageSize = decimal.NewFromInt(u.Get("average_message_size.0.value").Int())
	}

	if messageSize.GreaterThan(billableRequestSize) {
		monthlyMessages = calculateBillableRequests(messageSize, billableRequestSize, monthlyMessages)
	}

	apiRequestQuantities := calculateApiRequests(monthlyMessages, apiTierRequests, apiTierOneLimt, apiTierTwoLimit)

	apiTierOne := apiRequestQuantities["apiRequestTierOne"]
	apiTierTwo := apiRequestQuantities["apiRequestTierTwo"]

	return []*schema.CostComponent{
		{
			Name:            "Requests (first 1B message transfers)",
			Unit:            "requests",
			MonthlyQuantity: &apiTierOne,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonApiGateway"),
				ProductFamily: strPtr("WebSocket"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr("WebSocket Messages")},
					{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayMessage/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("0"),
			},
		},
		{
			Name:            "Requests (over 1B message transfers)",
			Unit:            "requests",
			MonthlyQuantity: &apiTierTwo,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonApiGateway"),
				ProductFamily: strPtr("WebSocket"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr("WebSocket Messages")},
					{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayMessage/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("1000000000"),
			},
		},
	}

}

func calculateApiRequests(requests decimal.Decimal, apiRequestTiers map[string]decimal.Decimal, apiTierOneLimit decimal.Decimal, apiTierTwoLimit decimal.Decimal) map[string]decimal.Decimal {

	if requests.GreaterThanOrEqual(apiTierOneLimit) {
		apiRequestTiers["apiRequestTierOne"] = apiTierOneLimit
	} else {
		apiRequestTiers["apiRequestTierOne"] = requests
		return apiRequestTiers
	}

	if requests.GreaterThanOrEqual(apiTierTwoLimit) {
		apiRequestTiers["apiRequestTierTwo"] = apiTierTwoLimit
	} else {
		apiRequestTiers["apiRequestTierTwo"] = requests.Sub(apiTierOneLimit)
		return apiRequestTiers
	}

	return apiRequestTiers
}

func calculateBillableRequests(requestSize decimal.Decimal, billableRequestSize decimal.Decimal, requests decimal.Decimal) decimal.Decimal {
	return requests.Mul(requestSize.Div(billableRequestSize).Ceil())
}


