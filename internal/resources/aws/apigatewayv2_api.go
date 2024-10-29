package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type APIGatewayV2API struct {
	Address               string
	Region                string
	ProtocolType          string
	MessageSizeKB         *int64 `infracost_usage:"message_size_kb"`
	MonthlyConnectionMins *int64 `infracost_usage:"monthly_connection_mins"`
	MonthlyRequests       *int64 `infracost_usage:"monthly_requests"`
	RequestSizeKB         *int64 `infracost_usage:"request_size_kb"`
	MonthlyMessages       *int64 `infracost_usage:"monthly_messages"`
}

func (r *APIGatewayV2API) CoreType() string {
	return "APIGatewayV2API"
}

func (r *APIGatewayV2API) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "message_size_kb", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_connection_mins", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "request_size_kb", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_messages", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *APIGatewayV2API) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *APIGatewayV2API) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent

	if strings.ToLower(r.ProtocolType) == "websocket" {
		costComponents = r.websocketAPICostComponent()
	}

	if strings.ToLower(r.ProtocolType) == "http" {
		costComponents = r.httpAPICostComponent()
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *APIGatewayV2API) httpAPICostComponent() []*schema.CostComponent {
	var monthlyRequests *decimal.Decimal
	requestSize := decimal.NewFromInt(512)

	billableRequestSize := decimal.NewFromInt(512)

	httpAPITiers := []int{300000000}

	costComponents := []*schema.CostComponent{}

	if r.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))

		if r.RequestSizeKB != nil {
			requestSize = decimal.NewFromInt(*r.RequestSizeKB)
		}

		if requestSize.GreaterThan(billableRequestSize) {
			monthlyRequests = calculateBillableRequests(&requestSize, &billableRequestSize, monthlyRequests)
		}

		apiRequestQuantities := usage.CalculateTierBuckets(*monthlyRequests, httpAPITiers)

		costComponents = append(costComponents, r.httpCostComponent("Requests (first 300M)", "0", &apiRequestQuantities[0]))

		if apiRequestQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.httpCostComponent("Requests (over 300M)", "300000000", &apiRequestQuantities[1]))
		}
	} else {
		costComponents = append(costComponents, r.httpCostComponent("Requests (first 300M)", "0", nil))
	}

	return costComponents
}

func (r *APIGatewayV2API) websocketAPICostComponent() []*schema.CostComponent {
	var monthlyMessages *decimal.Decimal
	var monthlyConnectionMinutes *decimal.Decimal

	messageSize := decimal.NewFromInt(32)

	billableRequestSize := decimal.NewFromInt(32)

	websocketAPITiers := []int{1000000000}

	costComponents := []*schema.CostComponent{}

	if r.MonthlyMessages != nil {
		monthlyMessages = decimalPtr(decimal.NewFromInt(*r.MonthlyMessages))

		if r.MessageSizeKB != nil {
			messageSize = decimal.NewFromInt(*r.MessageSizeKB)
		}

		if messageSize.GreaterThan(billableRequestSize) {
			monthlyMessages = calculateBillableRequests(&messageSize, &billableRequestSize, monthlyMessages)
		}

		apiRequestQuantities := usage.CalculateTierBuckets(*monthlyMessages, websocketAPITiers)

		costComponents = append(costComponents, r.websocketCostComponent("messages", "ApiGatewayMessage", "Messages (first 1B)", "0", &apiRequestQuantities[0]))

		if apiRequestQuantities[1].GreaterThan(decimal.NewFromInt(0)) {
			costComponents = append(costComponents, r.websocketCostComponent("messages", "ApiGatewayMessage", "Messages (over 1B)", "1000000000", &apiRequestQuantities[1]))
		}
	} else {
		costComponents = append(costComponents, r.websocketCostComponent("messages", "ApiGatewayMessage", "Messages (first 1B)", "0", nil))
	}

	if r.MonthlyConnectionMins != nil {
		monthlyConnectionMinutes = decimalPtr(decimal.NewFromInt(*r.MonthlyConnectionMins))
	}
	costComponents = append(costComponents, r.websocketCostComponent("minutes", "ApiGatewayMinute", "Connection duration", "0", monthlyConnectionMinutes))

	return costComponents
}

func calculateBillableRequests(requestSize *decimal.Decimal, billableRequestSize *decimal.Decimal, requests *decimal.Decimal) *decimal.Decimal {
	return decimalPtr(requests.Mul(requestSize.Div(*billableRequestSize).Ceil()))
}

func (r *APIGatewayV2API) httpCostComponent(displayName string, usageTier string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "1M requests",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonApiGateway"),
			ProductFamily: strPtr("API Calls"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/ApiGatewayHttpRequest/")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}

func (r *APIGatewayV2API) websocketCostComponent(unit string, usageType string, displayName string, usageTier string, monthlyQuantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            displayName,
		Unit:            "1M " + unit,
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonApiGateway"),
			ProductFamily: strPtr("WebSocket"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
		UsageBased: true,
	}
}
