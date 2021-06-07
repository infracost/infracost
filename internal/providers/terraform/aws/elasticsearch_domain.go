package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetElasticsearchDomainRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_elasticsearch_domain",
		RFunc: NewElasticsearchDomain,
	}
}

func NewElasticsearchDomain(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	defaultInstanceType := "m4.large.elasticsearch"

	instanceType := defaultInstanceType
	instanceCount := int64(1)
	dedicatedMasterEnabled := false
	ultrawarmEnabled := false
	ebsEnabled := false

	if d.Get("cluster_config.0.instance_type").Exists() {
		instanceType = d.Get("cluster_config.0.instance_type").String()
	}

	if d.Get("cluster_config.0.instance_count").Exists() {
		instanceCount = d.Get("cluster_config.0.instance_count").Int()
	}

	if d.Get("cluster_config.0.dedicated_master_enabled").Exists() {
		dedicatedMasterEnabled = d.Get("cluster_config.0.dedicated_master_enabled").Bool()
	}

	if d.Get("cluster_config.0.warm_enabled").Exists() {
		ultrawarmEnabled = d.Get("cluster_config.0.warm_enabled").Bool()
	}

	if d.Get("ebs_options.0.ebs_enabled").Exists() {
		ebsEnabled = d.Get("ebs_options.0.ebs_enabled").Bool()
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           fmt.Sprintf("Instance (on-demand, %s)", instanceType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(instanceCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Elastic Search Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
					{Key: "instanceType", Value: strPtr(fmt.Sprintf("/%s/i", instanceType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		},
	}

	if ebsEnabled {
		gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
		if d.Get("ebs_options.0.volume_size").Exists() {
			gbVal = decimal.NewFromFloat(d.Get("ebs_options.0.volume_size").Float())
		}

		ebsType := "gp2"
		if d.Get("ebs_options.0.volume_type").Exists() {
			ebsType = d.Get("ebs_options.0.volume_type").String()
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
			UnitMultiplier:  1,
			MonthlyQuantity: &gbVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Elastic Search Volume"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ES.+-Storage/")},
					{Key: "storageMedia", Value: strPtr(fmt.Sprintf("/%s/i", ebsFilter))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		})

		if strings.ToLower(ebsType) == "io1" {
			iopsVal := decimal.NewFromInt(1)
			if d.Get("ebs_options.0.iops").Exists() {
				iopsVal = decimal.NewFromFloat(d.Get("ebs_options.0.iops").Float())

				if iopsVal.LessThan(decimal.NewFromInt(1)) {
					iopsVal = decimal.NewFromInt(1)
				}
			}

			costComponents = append(costComponents, &schema.CostComponent{
				Name:            fmt.Sprintf("Storage IOPS (%s)", ebsType),
				Unit:            "IOPS",
				UnitMultiplier:  1,
				MonthlyQuantity: &iopsVal,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonES"),
					ProductFamily: strPtr("Elastic Search Volume"),
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

		if d.Get("cluster_config.0.dedicated_master_type").Type != gjson.Null {
			dedicatedMasterType = d.Get("cluster_config.0.dedicated_master_type").String()
		}

		if d.Get("cluster_config.0.dedicated_master_count").Type != gjson.Null {
			dedicatedMasterCount = d.Get("cluster_config.0.dedicated_master_count").Int()
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:           fmt.Sprintf("Dedicated master (on-demand, %s)", dedicatedMasterType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(dedicatedMasterCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Elastic Search Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
					{Key: "instanceType", Value: strPtr(fmt.Sprintf("/%s/i", dedicatedMasterType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		})
	}

	ultrawarmType := d.Get("cluster_config.0.warm_type").String()
	ultrawarmCount := d.Get("cluster_config.0.warm_count").Int()

	if ultrawarmEnabled && ultrawarmType != "" && ultrawarmCount > 0 {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:           fmt.Sprintf("UltraWarm instance (on-demand, %s)", ultrawarmType),
			Unit:           "hours",
			UnitMultiplier: 1,
			HourlyQuantity: decimalPtr(decimal.NewFromInt(ultrawarmCount)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonES"),
				ProductFamily: strPtr("Elastic Search Instance"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ESInstance/")},
					{Key: "instanceType", Value: strPtr(fmt.Sprintf("/%s/i", ultrawarmType))},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("on_demand"),
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
