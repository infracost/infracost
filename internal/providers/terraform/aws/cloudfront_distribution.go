package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetCloudfrontDistributionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudfront_distribution",
		RFunc: NewCloudfrontDistribution,
	}
}

func NewCloudfrontDistribution(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	return &schema.Resource{
		Name: d.Address,
		SubResources: []*schema.Resource{
			regionalDataOutToOrigin(u),
		},
	}
}

func regionalDataOutToOrigin(u *schema.ResourceData) *schema.Resource {
	resource := &schema.Resource{
		Name:           "Regional data transfer out to origin",
		CostComponents: []*schema.CostComponent{},
	}

	// regionMap structure is: aws grouped name -> [pricing region , usage data key]
	regionsMap := map[string][]string{
		"United States, Mexico, & Canada":      []string{"United States", "united_states_data_transfer_origin_gb"},
		"Europe & Israel":                      []string{"Europe", "europe_data_transfer_origin_gb"},
		"South Africa, Kenya, & Middle East":   []string{"South Africa", "south_africa_data_transfer_origin_gb"},
		"South America":                        []string{"South America", "south_america_data_transfer_origin_gb"},
		"Japan":                                []string{"Japan", "japan_data_transfer_origin_gb"},
		"Australia & New Zealand":              []string{"Australia", "australia_data_transfer_origin_gb"},
		"Hong Kong, Philippines, Asia Pacific": []string{"Asia Pacific", "asia_pacific_data_transfer_origin_gb"},
		"India":                                []string{"India", "india_data_transfer_origin_gb"},
	}

	for key, value := range regionsMap {
		awsRegion := key
		apiRegion := value[0]
		usageKey := value[1]
		var quantity *decimal.Decimal
		if u != nil && u.Get(usageKey).Exists() {
			quantity = decimalPtr(decimal.NewFromInt(u.Get(usageKey).Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "GB",
			UnitMultiplier:  1,
			MonthlyQuantity: quantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Service:    strPtr("AmazonCloudFront"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "transferType", Value: strPtr("CloudFront to Origin")},
					{Key: "fromLocation", Value: strPtr(apiRegion)},
				},
			},
		})
	}

	return resource
}
