package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetMSKClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_msk_cluster",
		RFunc: NewMskCluster,
	}
}

func NewMskCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	brokerNodes := decimal.NewFromInt(d.Get("number_of_broker_nodes").Int())
	instanceType := d.Get("broker_node_group_info.0.instance_type").String()
	ebsVolumeSize := decimal.NewFromInt(d.Get("broker_node_group_info.0.ebs_volume_size").Int()).Mul(brokerNodes)

	return &schema.Resource{
		Name: d.Address,
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
		},
	}
}
