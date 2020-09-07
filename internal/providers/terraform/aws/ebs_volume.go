package aws

import (
	"github.com/infracost/infracost/pkg/schema"

	"github.com/shopspring/decimal"
)

func NewEBSVolume(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {
	region := d.Get("region").String()

	volumeApiName := "gp2"
	if d.Get("type").Exists() {
		volumeApiName = d.Get("type").String()
	}

	gbVal := decimal.NewFromInt(int64(defaultVolumeSize))
	if d.Get("size").Exists() {
		gbVal = decimal.NewFromFloat(d.Get("size").Float())
	}

	iopsVal := decimal.Zero
	if d.Get("iops").Exists() {
		iopsVal = decimal.NewFromFloat(d.Get("iops").Float())
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: ebsVolumeCostComponents(region, volumeApiName, gbVal, iopsVal),
	}
}

func ebsVolumeCostComponents(region string, volumeApiName string, gbVal decimal.Decimal, iopsVal decimal.Decimal) []*schema.CostComponent {
	costComponents := []*schema.CostComponent{
		{
			Name:            "Storage",
			Unit:            "GB-months",
			MonthlyQuantity: &gbVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "volumeApiName", Value: strPtr(volumeApiName)},
				},
			},
		},
	}

	if volumeApiName == "io1" {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Storage IOPS",
			Unit:            "IOPS-months",
			MonthlyQuantity: &iopsVal,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEC2"),
				ProductFamily: strPtr("System Operation"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "volumeApiName", Value: strPtr(volumeApiName)},
					{Key: "usagetype", ValueRegex: strPtr("/EBS:VolumeP-IOPS.piops/")},
				},
			},
		})
	}

	return costComponents
}
