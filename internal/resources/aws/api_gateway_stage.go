package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type APIGatewayStage struct {
	Address          string
	Region           string
	CacheClusterSize float64
	CacheEnabled     bool
}

func (r *APIGatewayStage) CoreType() string {
	return "APIGatewayStage"
}

func (r *APIGatewayStage) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *APIGatewayStage) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *APIGatewayStage) BuildResource() *schema.Resource {
	if !r.CacheEnabled {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	region := r.Region

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Cache memory (%s GB)", decimal.NewFromFloat(r.CacheClusterSize)),
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonApiGateway"),
					ProductFamily: strPtr("Amazon API Gateway Cache"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "cacheMemorySizeGb", ValueRegex: strPtr(fmt.Sprintf("/%s/", decimal.NewFromFloat(r.CacheClusterSize)))},
					},
				},
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
