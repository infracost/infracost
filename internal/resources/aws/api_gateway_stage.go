package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type APIGatewayStage struct {
	Address          *string
	Region           *string
	CacheClusterSize *float64
}

var APIGatewayStageUsageSchema = []*schema.UsageItem{}

func (r *APIGatewayStage) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *APIGatewayStage) BuildResource() *schema.Resource {
	region := *r.Region

	cacheMemorySize := decimal.Zero

	if r.CacheClusterSize != nil {
		cacheMemorySize = decimal.NewFromFloat(*r.CacheClusterSize)
	}

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Cache memory (%s GB)", cacheMemorySize),
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonApiGateway"),
					ProductFamily: strPtr("Amazon API Gateway Cache"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "cacheMemorySizeGb", ValueRegex: strPtr(fmt.Sprintf("/%s/i", cacheMemorySize.String()))},
					},
				},
			},
		}, UsageSchema: APIGatewayStageUsageSchema,
	}
}
