package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type ElasticsearchDomain struct {
	Address              string
	Region               string
	ClusterInstanceType  string
	ClusterInstanceCount *int64 // If this is nil it will default to 1

	EBSEnabled    bool
	EBSVolumeType string
	EBSVolumeSize *float64 // if this is nil it will default to 8
	EBSIOPS       *float64 // if this is nil it will default to 1

	ClusterDedicatedMasterEnabled bool
	ClusterDedicatedMasterType    string
	ClusterDedicatedMasterCount   *int64 // if this is nil it will default to 3

	ClusterWarmEnabled bool
	ClusterWarmType    string
	ClusterWarmCount   *int64
}

var ElasticsearchDomainUsageSchema = []*schema.UsageItem{}

func (r *ElasticsearchDomain) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ElasticsearchDomain) BuildResource() *schema.Resource {
	defaultClusterInstanceType := "m4.large.elasticsearch"

	instanceType := defaultClusterInstanceType
	if r.ClusterInstanceType != "" {
		instanceType = r.ClusterInstanceType
	}

	instanceCount := int64(1)
	if r.ClusterInstanceCount != nil {
		instanceCount = *r.ClusterInstanceCount
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Instance (on-demand, %s)", instanceType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(instanceCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Amazon OpenSearch Service Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
					{Key: "instanceType", Value: r.opensearchifyClusterInstanceType(instanceType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if r.EBSEnabled {
		gbVal := decimal.NewFromFloat(float64(defaultVolumeSize))
		if r.EBSVolumeSize != nil {
			gbVal = decimal.NewFromFloat(*r.EBSVolumeSize)
		}

		ebsType := "gp2"
		if r.EBSVolumeType != "" {
			ebsType = r.EBSVolumeType
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
				Region:        strPtr(r.Region),
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
			if r.EBSIOPS != nil {
				iopsVal = decimal.NewFromFloat(*r.EBSIOPS)

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
					Region:        strPtr(r.Region),
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

	if r.ClusterDedicatedMasterEnabled {
		dedicatedMasterType := defaultClusterInstanceType
		if r.ClusterDedicatedMasterType != "" {
			dedicatedMasterType = r.ClusterDedicatedMasterType
		}

		dedicatedMasterCount := int64(3)
		if r.ClusterDedicatedMasterCount != nil {
			dedicatedMasterCount = *r.ClusterDedicatedMasterCount
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:           fmt.Sprintf("Dedicated master (on-demand, %s)", dedicatedMasterType),
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(dedicatedMasterCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.Region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Amazon OpenSearch Service Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
					{Key: "instanceType", Value: r.opensearchifyClusterInstanceType(dedicatedMasterType)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		})
	}

	if r.ClusterWarmEnabled && r.ClusterWarmType != "" {
		clusterWarmCount := int64(0)
		if r.ClusterWarmCount != nil {
			clusterWarmCount = *r.ClusterWarmCount
		}

		if clusterWarmCount > 0 {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:           fmt.Sprintf("UltraWarm instance (on-demand, %s)", r.ClusterWarmType),
				Unit:           "hours",
				UnitMultiplier: decimal.NewFromInt(1),
				HourlyQuantity: decimalPtr(decimal.NewFromInt(clusterWarmCount)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("AmazonES"),
					ProductFamily: strPtr("Amazon OpenSearch Service Instance"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
						{Key: "instanceType", Value: r.opensearchifyClusterInstanceType(r.ClusterWarmType)},
					},
				},
				PriceFilter: &schema.PriceFilter{
					PurchaseOption: strPtr("on_demand"),
				},
			})
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    ElasticsearchDomainUsageSchema,
	}
}

func (r *ElasticsearchDomain) opensearchifyClusterInstanceType(instanceType string) *string {
	s := strings.Replace(instanceType, ".elasticsearch", ".search", 1)
	return &s
}
