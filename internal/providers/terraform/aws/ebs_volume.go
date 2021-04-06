package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

func GetEBSVolumeRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_ebs_volume",
		RFunc: NewEBSVolume,
	}
}

func NewEBSVolume(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	volumeAPIName := "gp2"
	if d.Get("type").Exists() {
		volumeAPIName = d.Get("type").String()
	}

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
	if d.Get("size").Exists() {
		gbVal = decimal.NewFromFloat(d.Get("size").Float())
	}

	iopsVal := decimal.Zero
	if d.Get("iops").Exists() {
		iopsVal = decimal.NewFromFloat(d.Get("iops").Float())
	}

	var throughputVal *decimal.Decimal
	if d.Get("throughput").Exists() {
		throughputVal = decimalPtr(decimal.NewFromInt(d.Get("throughput").Int()))
	}

	var monthlyIORequests *decimal.Decimal
	if u != nil && u.Get("monthly_standard_io_requests").Exists() {
		monthlyIORequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_standard_io_requests").Int()))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: ebsVolumeCostComponents(region, volumeAPIName, throughputVal, gbVal, iopsVal, monthlyIORequests),
	}
}

func ebsVolumeCostComponents(region string, volumeAPIName string, throughputVal *decimal.Decimal, gbVal decimal.Decimal, iopsVal decimal.Decimal, ioRequests *decimal.Decimal) []*schema.CostComponent {
	if volumeAPIName == "" {
		volumeAPIName = "gp2"
	}

	var name, usageType string
	switch volumeAPIName {
	case "standard":
		name = "Storage (magnetic)"
		usageType = "EBS:VolumeIOUsage"
	case "io1":
		name = "Storage (provisioned IOPS SSD, io1)"
		usageType = "EBS:VolumeP-IOPS.piops"
	case "io2":
		name = "Storage (provisioned IOPS SSD, io2)"
		usageType = "EBS:VolumeP-IOPS.io2$"
	case "st1":
		name = "Storage (throughput optimized HDD, st1)"
	case "sc1":
		name = "Storage (cold HDD, sc1)"
	case "gp3":
		name = "Storage (general purpose SSD, gp3)"
	default:
		name = "Storage (general purpose SSD, gp2)"
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            name,
			Unit:            "GB-months",
			UnitMultiplier:  1,
			MonthlyQuantity: &gbVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "volumeApiName", Value: strPtr(volumeAPIName)},
				},
			},
		},
	}

	if volumeAPIName == "io1" || volumeAPIName == "io2" {
		costComponents = append(costComponents, ebsProvisionedIops(region, volumeAPIName, usageType, &iopsVal))

	}

	if volumeAPIName == "standard" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "I/O requests",
			Unit:            "request",
			UnitMultiplier:  1000000,
			MonthlyQuantity: ioRequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "volumeApiName", Value: strPtr(volumeAPIName)},
					{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", usageType))},
				},
			},
		})
	}

	if volumeAPIName == "gp3" {
		if throughputVal != nil {
			if throughputVal.GreaterThan(decimal.NewFromInt(125)) {
				throughputVal = decimalPtr(throughputVal.Sub(decimal.NewFromInt(125)))
				costComponents = append(costComponents, &schema.CostComponent{
					Name:            "Provisioned throughput",
					Unit:            "Mbps-months",
					UnitMultiplier:  1,
					MonthlyQuantity: throughputVal,
					ProductFilter: &schema.ProductFilter{
						VendorName:    strPtr("aws"),
						Region:        strPtr(region),
						Service:       strPtr("AmazonEC2"),
						ProductFamily: strPtr("Provisioned Throughput"),
						AttributeFilters: []*schema.AttributeFilter{
							{Key: "volumeApiName", Value: strPtr(volumeAPIName)},
							{Key: "usagetype", ValueRegex: strPtr("/VolumeP-Throughput.gp3/")},
						},
					},
					PriceFilter: &schema.PriceFilter{
						Unit: strPtr("MiBps-Mo"),
					},
				})

			}
		}
		if iopsVal.GreaterThan((decimal.NewFromInt(3000))) {
			iopsVal = iopsVal.Sub(decimal.NewFromInt(3000))
			costComponents = append(costComponents, ebsProvisionedIops(region, volumeAPIName, "VolumeP-IOPS.gp3", &iopsVal))
		}

	}

	return costComponents
}
func ebsProvisionedIops(region string, volumeAPIName string, usageType string, iopsVal *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Provisioned IOPS",
		Unit:            "IOPS-months",
		UnitMultiplier:  1,
		MonthlyQuantity: iopsVal,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr("AmazonEC2"),
			ProductFamily: strPtr("System Operation"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "volumeApiName", Value: strPtr(volumeAPIName)},
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", usageType))},
			},
		},
	}
}
