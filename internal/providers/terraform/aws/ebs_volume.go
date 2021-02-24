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

	var monthlyIORequests *decimal.Decimal
	if u != nil && u.Get("monthly_input_output_operations").Exists() {
		monthlyIORequests = decimalPtr(decimal.NewFromInt(u.Get("monthly_input_output_operations").Int()))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: ebsVolumeCostComponents(region, volumeAPIName, gbVal, iopsVal, monthlyIORequests),
	}
}

func ebsVolumeCostComponents(region string, volumeAPIName string, gbVal decimal.Decimal, iopsVal decimal.Decimal, IORequests *decimal.Decimal) []*schema.CostComponent {
	if volumeAPIName == "" {
		volumeAPIName = "gp2"
	}

	var name, usageType string
	switch volumeAPIName {
	case "standard":
		name = "Magnetic storage"
		usageType = "EBS:VolumeIOUsage"
	case "io1":
		name = "Provisioned IOPS SSD storage (io1)"
		usageType = "EBS:VolumeP-IOPS.piops"
	case "io2":
		name = "Provisioned IOPS SSD storage (io2)"
		usageType = "EBS:VolumeP-IOPS.io2$"
	case "st1":
		name = "Throughput Optimized HDD storage (st1)"
	case "sc1":
		name = "Cold HDD storage (sc1)"
	default:
		name = "General Purpose SSD storage (gp2)"
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
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Provisioned IOPS",
			Unit:            "IOPS-months",
			UnitMultiplier:  1,
			MonthlyQuantity: &iopsVal,
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

	if volumeAPIName == "standard" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "I/O requests",
			Unit:            "request",
			UnitMultiplier:  1000000,
			MonthlyQuantity: IORequests,
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

	return costComponents
}
