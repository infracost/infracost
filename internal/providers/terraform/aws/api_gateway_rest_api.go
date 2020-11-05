package aws

import (
    "github.com/infracost/infracost/internal/schema"
    "github.com/shopspring/decimal"
)

func GetApiGatewayRestApiRegistryItem() *schema.RegistryItem {
    return &schema.RegistryItem{
        Name:  "aws_api_gateway_rest_api",
        RFunc: NewApiGatewayRestApi,
    }
}

func NewApiGatewayRestApi(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
    region := d.Get("region").String()

    var apiTierRequests = map[string]decimal.Decimal{
        "tierOne":   decimal.Zero,
        "tierTwo":   decimal.Zero,
        "tierThree": decimal.Zero,
        "tierFour":  decimal.Zero,
    }

    monthlyRequests := decimal.Zero

    if u != nil && u.Get("monthly_requests.0.value").Exists() {
        monthlyRequests = decimal.NewFromInt(u.Get("monthly_requests.0.value").Int())
    }

    apiRequestQuantities := calculateApiRequests(monthlyRequests, apiTierRequests)

    tierOne   := apiRequestQuantities["tierOne"]
    tierTwo   := apiRequestQuantities["tierTwo"]
    tierThree := apiRequestQuantities["tierThree"]
    tierFour  := apiRequestQuantities["tierFour"]

    return &schema.Resource{
        Name: d.Address,
        CostComponents: []*schema.CostComponent{
            {
                Name:            "First 333 Million Requests",
                Unit:            "requests",
                MonthlyQuantity: &tierOne,
                ProductFilter: &schema.ProductFilter{
                    VendorName:    strPtr("aws"),
                    Region:        strPtr(region),
                    Service:       strPtr("AmazonApiGateway"),
                    ProductFamily: strPtr("API Calls"),
                    AttributeFilters: []*schema.AttributeFilter{
                        {Key: "description", Value: strPtr("API calls received")},
                        {Key: "usagetype", ValueRegex: strPtr("/ApiGatewayRequest/")},
                        {Key: "operation", Value: strPtr("ApiGatewayRequest")},
                    },
                },
                PriceFilter: &schema.PriceFilter{
                    StartUsageAmount: strPtr("0"),
                    EndUsageAmount: strPtr("333000000"),
                },
            },
            {
                Name:            "Next 667 Million Requests",
                Unit:            "requests",
                MonthlyQuantity: &tierTwo,
                ProductFilter: &schema.ProductFilter{
                    VendorName:    strPtr("aws"),
                    Region:        strPtr(region),
                    Service:       strPtr("AmazonApiGateway"),
                    ProductFamily: strPtr("API Calls"),
                    AttributeFilters: []*schema.AttributeFilter{
                        {Key: "description", Value: strPtr("API calls received")},
                        {Key: "usagetype", ValueRegex: strPtr("/ApiGatewayRequest/")},
                        {Key: "operation", Value: strPtr("ApiGatewayRequest")},
                    },
                },
                PriceFilter: &schema.PriceFilter{
                    StartUsageAmount: strPtr("333000000"),
                    EndUsageAmount: strPtr("1000000000"),
                },
            },
            {
                Name:            "Next 19 Billion Requests",
                Unit:            "requests",
                MonthlyQuantity: &tierThree,
                ProductFilter: &schema.ProductFilter{
                    VendorName:    strPtr("aws"),
                    Region:        strPtr(region),
                    Service:       strPtr("AmazonApiGateway"),
                    ProductFamily: strPtr("API Calls"),
                    AttributeFilters: []*schema.AttributeFilter{
                        {Key: "description", Value: strPtr("API calls received")},
                        {Key: "usagetype", ValueRegex: strPtr("/ApiGatewayRequest/")},
                        {Key: "operation", Value: strPtr("ApiGatewayRequest")},
                    },
                },
                PriceFilter: &schema.PriceFilter{
                    StartUsageAmount: strPtr("1000000000"),
                    EndUsageAmount: strPtr("20000000000"),
                },
            },
            {
                Name:            "Over 20 Billion Requests",
                Unit:            "requests",
                MonthlyQuantity: &tierFour,
                ProductFilter: &schema.ProductFilter{
                    VendorName:    strPtr("aws"),
                    Region:        strPtr(region),
                    Service:       strPtr("AmazonApiGateway"),
                    ProductFamily: strPtr("API Calls"),
                    AttributeFilters: []*schema.AttributeFilter{
                        {Key: "description", Value: strPtr("API calls received")},
                        {Key: "usagetype", ValueRegex: strPtr("/ApiGatewayRequest/")},
                        {Key: "operation", Value: strPtr("ApiGatewayRequest")},
                    },
                },
                PriceFilter: &schema.PriceFilter{
                    StartUsageAmount: strPtr("20000000000"),
                },
            },
        },
    }
}

func calculateApiRequests(requests decimal.Decimal, tiers map[string]decimal.Decimal) map[string]decimal.Decimal {

    // API gateway charging tiers
    apiTierOneLimit := decimal.NewFromInt(333000000)
    apiTierTwoLimit := decimal.NewFromInt(667000000)
    apiTierThreeLimit := decimal.NewFromInt(20000000000)
    apiTierFourLimit := decimal.NewFromInt(21000000000)

    if requests.GreaterThanOrEqual(apiTierOneLimit) {
        tiers["tierOne"] = apiTierOneLimit
    } else {
        tiers["tierOne"] = requests
        return tiers
    }

    if requests.GreaterThanOrEqual(apiTierTwoLimit) {
        tiers["tierTwo"]= apiTierTwoLimit
    } else {
        tiers["tierTwo"] = requests.Sub(apiTierOneLimit)
        return tiers
    }

    if requests.GreaterThanOrEqual(apiTierThreeLimit) {
        tiers["tierThree"]= apiTierThreeLimit
    } else {
        tiers["tierThree"] = requests.Sub(apiTierTwoLimit.Add(apiTierOneLimit))
        return tiers
    }

    if requests.GreaterThanOrEqual(apiTierFourLimit) {
        tiers["tierFour"] = requests.Sub(apiTierFourLimit)
        return tiers
    }

    return tiers
}

