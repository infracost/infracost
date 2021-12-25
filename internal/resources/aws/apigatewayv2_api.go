package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"strings"
)

type APIGatewayv2Api struct {
	Address               *string
	ProtocolType          *string
	Region                *string
	MonthlyConnectionMins *int64 `infracost_usage:"monthly_connection_mins"`
	MonthlyRequests       *int64 `infracost_usage:"monthly_requests"`
	RequestSizeKb         *int64 `infracost_usage:"request_size_kb"`
	MonthlyMessages       *int64 `infracost_usage:"monthly_messages"`
	MessageSizeKb         *int64 `infracost_usage:"message_size_kb"`
}

var APIGatewayv2ApiUsageSchema = []*schema.UsageItem{{Key: "monthly_connection_mins", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_requests", ValueType: schema.Int64, DefaultValue: 0}, {Key: "request_size_kb", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_messages", ValueType: schema.Int64, DefaultValue: 0}, {Key: "message_size_kb", ValueType: schema.Int64, DefaultValue: 0}}

func (r *APIGatewayv2Api) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *APIGatewayv2Api) BuildResource() *schema.Resource {
	var costComponents []*schema.CostComponent

	protocolType := *r.ProtocolType

	if strings.ToLower(protocolType) == "websocket" {
		costComponents = websocketAPICostComponent(r)
	}

	if strings.ToLower(protocolType) == "http" {
		costComponents = httpAPICostComponent(r)
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: APIGatewayv2ApiUsageSchema,
	}
}

func httpAPICostComponent(r *APIGatewayv2Api,) []*schema.CostComponent {
	region := *r.Region
	var monthlyRequests *decimal.Decimal
	requestSize := decimal.NewFromInt(512)

	billableRequestSize := decimal.NewFromInt(512)

	httpAPITiers := []int{300000000}

	costComponents := []*schema.CostComponent{}

	if r != nil && r.MonthlyRequests != nil {
		monthlyRequests = decimalPtr(decimal.NewFromInt(*r.MonthlyRequests))

		if r.RequestSizeKb != nil {
			requestSize = decimal.NewFromInt(*r.RequestSizeKb)
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

func websocketAPICostComponent(r *APIGatewayv2Api,) []*schema.CostComponent {
	region := *r.Region
	var monthlyMessages *decimal.Decimal
	var monthlyConnectionMinutes *decimal.Decimal

	messageSize := decimal.NewFromInt(32)

	billableRequestSize := decimal.NewFromInt(32)

	websocketAPITiers := []int{1000000000}

	costComponents := []*schema.CostComponent{}

	if r != nil && r.MonthlyMessages != nil {
		monthlyMessages = decimalPtr(decimal.NewFromInt(*r.MonthlyMessages))

		if r.MessageSizeKb != nil {
			messageSize = decimal.NewFromInt(*r.MessageSizeKb)
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

	if r != nil && r.MonthlyConnectionMins != nil {
		monthlyConnectionMinutes = decimalPtr(decimal.NewFromInt(*r.MonthlyConnectionMins))
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
		UnitMultiplier:  decimal.NewFromInt(1000000),
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
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: monthlyQuantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonApiGateway"),
			ProductFamily: strPtr("WebSocket"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", usageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(usageTier),
		},
	}
}
