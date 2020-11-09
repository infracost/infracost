package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetAPIGatewayv2ApiRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name: "aws_apigatewayv2_api",
		Notes: []string{
			"WebSocket Connection minutes is not yet supported",
		},
		RFunc: NewAPIGatewayv2Api,
	}
}

func NewAPIGatewayv2Api(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	var costComponents []*schema.CostComponent

	protocolType := d.Get("protocol_type").String()

	if protocolType == "WEBSOCKET" {
		costComponents = websocketAPICostComponent(d, u)
	}

	if protocolType == "HTTP" {
		costComponents = httpAPICostComponent(d, u)
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func httpAPICostComponent(d *schema.ResourceData, u *schema.ResourceData) []*schema.CostComponent {
	region := d.Get("region").String()
	monthlyRequests := decimal.Zero
	requestSize := decimal.NewFromInt(512)

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

	if u != nil && u.Get("average_request_size.0.value").Exists() {
		requestSize = decimal.NewFromInt(u.Get("average_request_size.0.value").Int())
	}

	if requestSize.GreaterThan(billableRequestSize) {
		monthlyRequests = calculateBillableRequests(requestSize, billableRequestSize, monthlyRequests)
	}

	apiRequestQuantities := calculateAPIV2Requests(monthlyRequests, apiTierRequests, apiTierOneLimit, apiTierTwoLimit)

	apiTierOne := apiRequestQuantities["apiRequestTierOne"]
	apiTierTwo := apiRequestQuantities["apiRequestTierTwo"]

	CostComponent := []*schema.CostComponent{
		{
			Name:            "Requests (first 300M)",
			Unit:            "requests",
			UnitMultiplier:  1000000,
			MonthlyQuantity: &apiTierOne,
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
				StartUsageAmount: strPtr("0"),
			},
		},
	}

	if apiTierTwo.GreaterThan(decimal.NewFromInt(0)) {
		CostComponent = append(CostComponent, &schema.CostComponent{
			Name:            "Requests (over 300M)",
			Unit:            "requests",
			UnitMultiplier:  1000000,
			MonthlyQuantity: &apiTierTwo,
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
				StartUsageAmount: strPtr("300000000"),
			},
		})
	}

	return CostComponent
}

func websocketAPICostComponent(d *schema.ResourceData, u *schema.ResourceData) []*schema.CostComponent {
	region := d.Get("region").String()
	monthlyMessages := decimal.Zero
	messageSize := decimal.NewFromInt(32)

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

	apiRequestQuantities := calculateAPIV2Requests(monthlyMessages, apiTierRequests, apiTierOneLimt, apiTierTwoLimit)

	apiTierOne := apiRequestQuantities["apiRequestTierOne"]
	apiTierTwo := apiRequestQuantities["apiRequestTierTwo"]

	costComponent := []*schema.CostComponent{
		{
			Name:            "Messages (first 1B)",
			Unit:            "messages",
			UnitMultiplier:  1000000,
			MonthlyQuantity: &apiTierOne,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonApiGateway"),
				ProductFamily: strPtr("WebSocket"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayMessage/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("0"),
			},
		},
	}

	if apiTierTwo.GreaterThan(decimal.NewFromInt(0)) {
		costComponent = append(costComponent, &schema.CostComponent{
			Name:            "Messages (over 1B)",
			Unit:            "messages",
			UnitMultiplier:  1000000,
			MonthlyQuantity: &apiTierTwo,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonApiGateway"),
				ProductFamily: strPtr("WebSocket"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayMessage/")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("1000000000"),
			},
		})
	}

	return costComponent

}

func calculateAPIV2Requests(requests decimal.Decimal, apiRequestTiers map[string]decimal.Decimal, apiTierOneLimit decimal.Decimal, apiTierTwoLimit decimal.Decimal) map[string]decimal.Decimal {

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
