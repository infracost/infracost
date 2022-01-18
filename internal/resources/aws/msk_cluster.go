package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type MskCluster struct {
	Address                           *string
	Region                            *string
	NumberOfBrokerNodes               *int64
	BrokerNodeGroupInfo0InstanceType  *string
	BrokerNodeGroupInfo0EbsVolumeSize *int64
}

var MskClusterUsageSchema = []*schema.UsageItem{}

func (r *MskCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MskCluster) BuildResource() *schema.Resource {
	region := *r.Region

	brokerNodes := decimal.NewFromInt(*r.NumberOfBrokerNodes)
	instanceType := *r.BrokerNodeGroupInfo0InstanceType
	ebsVolumeSize := decimal.NewFromInt(*r.BrokerNodeGroupInfo0EbsVolumeSize).Mul(brokerNodes)

	return &schema.Resource{
		Name: *r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Instance (%s)", instanceType),
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: &brokerNodes,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonMSK"),
					ProductFamily: strPtr("Managed Streaming for Apache Kafka (MSK)"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", instanceType))},
						{Key: "locationType", Value: strPtr("AWS Region")},
					},
				},
			},
			{
				Name:            "Storage",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(ebsVolumeSize),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonMSK"),
					ProductFamily: strPtr("Managed Streaming for Apache Kafka (MSK)"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "storageFamily", Value: strPtr("GP2")},
					},
				},
			},
		}, UsageSchema: MskClusterUsageSchema,
	}
}
