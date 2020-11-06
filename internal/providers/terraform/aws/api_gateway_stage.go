package aws

import (
    "fmt"
    "github.com/infracost/infracost/internal/schema"
    "github.com/shopspring/decimal"
)

func GetApiGatewayStageRegistryItem() *schema.RegistryItem {
    return &schema.RegistryItem{
        Name:  "aws_api_gateway_stage",
        RFunc: NewApiGatewayStage,
    }
}

func NewApiGatewayStage(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
    region := d.Get("region").String()

    cacheMemorySize := decimal.Zero

    if d.Get("cache_cluster_size").Exists() {
        cacheMemorySize = decimal.NewFromFloat(d.Get("cache_cluster_size").Float())
    }

    return &schema.Resource{
        Name: d.Address,
        CostComponents: []*schema.CostComponent{
            {
                Name:           fmt.Sprintf("Cache Memory Size %s(GB)", cacheMemorySize),
                Unit:           "hours",
                HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
                ProductFilter: &schema.ProductFilter{
                    VendorName:    strPtr("aws"),
                    Region:        strPtr(region),
                    Service:       strPtr("AmazonApiGateway"),
                    ProductFamily: strPtr("Amazon API Gateway Cache"),
                    AttributeFilters: []*schema.AttributeFilter{
                        {Key: "cacheMemorySizeGb", Value: strPtr(cacheMemorySize.String())},
                    },
                },
            },
        },
    }
}