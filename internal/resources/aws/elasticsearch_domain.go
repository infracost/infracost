package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type ElasticsearchDomain struct {
	Address                              *string
	Region                               *string
	ClusterConfig0InstanceType           *string
	ClusterConfig0InstanceCount          *int64
	ClusterConfig0WarmType               *string
	ClusterConfig0DedicatedMasterEnabled *bool
	ClusterConfig0WarmEnabled            *bool
	EbsOptions0EbsEnabled                *bool
	EbsOptions0VolumeSize                *float64
	EbsOptions0VolumeType                *string
	EbsOptions0Iops                      *float64
	ClusterConfig0DedicatedMasterType    *string
	ClusterConfig0DedicatedMasterCount   *int64
	ClusterConfig0WarmCount              *int64
}

var ElasticsearchDomainUsageSchema = []*schema.UsageItem{}

func (r *ElasticsearchDomain) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ElasticsearchDomain) BuildResource() *schema.Resource {
	region := *r.Region

	defaultInstanceType := "m4.large.elasticsearch"

	instanceType := defaultInstanceType
	instanceCount := int64(1)
	dedicatedMasterEnabled := false
	ultrawarmEnabled := false
	ebsEnabled := false

	if r.ClusterConfig0InstanceType != nil {
		instanceType = *r.ClusterConfig0InstanceType
	}

	if r.ClusterConfig0InstanceCount != nil {
		instanceCount = *r.ClusterConfig0InstanceCount
	}

	if r.ClusterConfig0DedicatedMasterEnabled != nil {
		dedicatedMasterEnabled = *r.ClusterConfig0DedicatedMasterEnabled
	}

	if r.ClusterConfig0WarmEnabled != nil {
		ultrawarmEnabled = *r.ClusterConfig0WarmEnabled
	}

	if r.EbsOptions0EbsEnabled != nil {
		ebsEnabled = *r.EbsOptions0EbsEnabled
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Instance (on-demand, %s)", instanceType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(instanceCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Amazon OpenSearch Service Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
					{Key: "instanceType", Value: opensearchifyInstanceType(instanceType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if ebsEnabled {
		gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
		if r.EbsOptions0VolumeSize != nil {
			gbVal = decimal.NewFromFloat(*r.EbsOptions0VolumeSize)
		}

		ebsType := "gp2"
		if r.EbsOptions0VolumeType != nil {
			ebsType = *r.EbsOptions0VolumeType
		}

		ebsTypeMap := map[string]string{
			"gp2":      "GP2",
			"io1":      "PIOPS-Storage",
			"standard": "Magnetic",
		}

		ebsFilter := "gp2"
		if val, ok := ebsTypeMap[ebsType]; ok {
			ebsFilter = val
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            fmt.Sprintf("Storage (%s)", ebsType),
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: &gbVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Amazon OpenSearch Service Volume"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ES.+-Storage/")},
					{Key: "storageMedia", Value: strPtr(ebsFilter)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		})

		if strings.ToLower(ebsType) == "io1" {
			iopsVal := decimal.NewFromInt(1)
			if r.EbsOptions0Iops != nil {
				iopsVal = decimal.NewFromFloat(*r.EbsOptions0Iops)

				if iopsVal.LessThan(decimal.NewFromInt(1)) {
					iopsVal = decimal.NewFromInt(1)
				}
			}

			costComponents = append(costComponents, &schema.CostComponent{
				Name:            fmt.Sprintf("Storage IOPS (%s)", ebsType),
				Unit:            "IOPS",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: &iopsVal,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonES"),
					ProductFamily: strPtr("Amazon OpenSearch Service Volume"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ES:PIOPS/")},
						{Key: "storageMedia", Value: strPtr("PIOPS")},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("on_demand"),
				},
			})
		}
	}

	if dedicatedMasterEnabled {
		dedicatedMasterType := defaultInstanceType
		dedicatedMasterCount := int64(3)

		if r.ClusterConfig0DedicatedMasterType != nil {
			dedicatedMasterType = *r.ClusterConfig0DedicatedMasterType
		}

		if r.ClusterConfig0DedicatedMasterCount != nil {
			dedicatedMasterCount = *r.ClusterConfig0DedicatedMasterCount
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:           fmt.Sprintf("Dedicated master (on-demand, %s)", dedicatedMasterType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(dedicatedMasterCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Amazon OpenSearch Service Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
					{Key: "instanceType", Value: opensearchifyInstanceType(dedicatedMasterType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		})
	}

	ultrawarmType := *r.ClusterConfig0WarmType
	ultrawarmCount := *r.ClusterConfig0WarmCount

	if ultrawarmEnabled && ultrawarmType != "" && ultrawarmCount > 0 {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           fmt.Sprintf("UltraWarm instance (on-demand, %s)", ultrawarmType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(ultrawarmCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Amazon OpenSearch Service Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
					{Key: "instanceType", Value: opensearchifyInstanceType(ultrawarmType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		})
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: costComponents, UsageSchema: ElasticsearchDomainUsageSchema,
	}
}

func opensearchifyInstanceType(instanceType string) *string {
	s := strings.Replace(instanceType, ".elasticsearch", ".search", 1)
	return &s
}
