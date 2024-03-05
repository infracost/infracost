package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"

	"github.com/shopspring/decimal"
)

type MSKCluster struct {
	Address                 string
	Region                  string
	BrokerNodes             int64
	BrokerNodeInstanceType  string
	BrokerNodeEBSVolumeSize int64

	// "optional" args, that may be empty depending on the resource config
	AppAutoscalingTarget []*AppAutoscalingTarget
}

func (r *MSKCluster) CoreType() string {
	return "MSKCluster"
}

func (r *MSKCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *MSKCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MSKCluster) BuildResource() *schema.Resource {
	ebsVolumeSize := r.BrokerNodeEBSVolumeSize
	ebsAutoscaleSuffix := ""

	for _, target := range r.AppAutoscalingTarget {
		if target.ScalableDimension == "kafka:broker-storage:VolumeSize" {
			ebsAutoscaleSuffix = " (autoscaling)"
			if target.Capacity != nil {
				ebsVolumeSize = *target.Capacity
			} else {
				ebsVolumeSize = target.MinCapacity
			}
		}
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:           fmt.Sprintf("Instance (%s)", r.BrokerNodeInstanceType),
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(r.BrokerNodes)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonMSK"),
					ProductFamily: strPtr("Managed Streaming for Apache Kafka (MSK)"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/i", r.BrokerNodeInstanceType))},
						{Key: "locationType", Value: strPtr("AWS Region")},
					},
				},
			},
			{
				Name:            "Storage" + ebsAutoscaleSuffix,
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(ebsVolumeSize * r.BrokerNodes)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonMSK"),
					ProductFamily: strPtr("Managed Streaming for Apache Kafka (MSK)"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "storageFamily", Value: strPtr("GP2")},
					},
				},
				UsageBased: true,
			},
		},
		UsageSchema: r.UsageSchema(),
	}
}
