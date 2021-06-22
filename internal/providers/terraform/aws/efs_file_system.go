package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetEFSFileSystemRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_efs_file_system",
		RFunc: NewEFSFileSystem,
		Notes: []string{"In Terraform config the availability_zone_name field is required in this resource for correct results"},
	}
}

func NewEFSFileSystem(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	var availabilityZoneName string
	costComponents := make([]*schema.CostComponent, 0)
	var gbStorage, infrequentAccessGbStorage *decimal.Decimal

	if u != nil && u.Get("infrequent_access_storage_gb").Type != gjson.Null {
		infrequentAccessGbStorage = decimalPtr(decimal.NewFromFloat(u.Get("infrequent_access_storage_gb").Float()))
	}
	if u != nil && u.Get("storage_gb").Type != gjson.Null {
		gbStorage = decimalPtr(decimal.NewFromFloat(u.Get("storage_gb").Float()))
	}

	if d.Get("availability_zone_name").Type != gjson.Null {
		availabilityZoneName = d.Get("availability_zone_name").String()
		if strings.HasPrefix(strings.ToLower(availabilityZoneName), "standard") {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Storage (standard)",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: gbStorage,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEFS"),
					ProductFamily: strPtr("Storage"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-TimedStorage-ByteHrs/")},
					},
				},
			})
		}
		if strings.HasPrefix(strings.ToLower(availabilityZoneName), "one") {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Storage (one zone)",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: gbStorage,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEFS"),
					ProductFamily: strPtr("Storage"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-TimedStorage-Z-ByteHrs/")},
					},
				},
			})
		}
	}

	if d.Get("provisioned_throughput_in_mibps").Exists() && d.Get("provisioned_throughput_in_mibps").Type != gjson.Null {
		throughput := decimal.NewFromFloat(d.Get("provisioned_throughput_in_mibps").Float())
		provisionedThroughput := calculateProvisionedThroughput(gbStorage, throughput)

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Provisioned throughput",
			Unit:            "MBps",
			UnitMultiplier:  1,
			MonthlyQuantity: provisionedThroughput,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEFS"),
				ProductFamily: strPtr("Provisioned Throughput"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/ProvisionedTP-MiBpsHrs/")},
				},
			},
		})
	}

	if len(d.Get("lifecycle_policy").Array()) > 0 {

		var infrequentAccessReadGbRequests *decimal.Decimal
		if u != nil && u.Get("monthly_infrequent_access_read_gb").Exists() {
			infrequentAccessReadGbRequests = decimalPtr(decimal.NewFromFloat(u.Get("monthly_infrequent_access_read_gb").Float()))
		}

		var infrequentAccessWriteGbRequests *decimal.Decimal
		if u != nil && u.Get("monthly_infrequent_access_write_gb").Exists() {
			infrequentAccessWriteGbRequests = decimalPtr(decimal.NewFromFloat(u.Get("monthly_infrequent_access_write_gb").Float()))
		}

		if strings.HasPrefix(strings.ToLower(availabilityZoneName), "standard") {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Storage (standard, infrequent access)",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: infrequentAccessGbStorage,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEFS"),
					ProductFamily: strPtr("Storage"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/-IATimedStorage-ByteHrs/")},
					},
				},
			})
		}
		if strings.HasPrefix(strings.ToLower(availabilityZoneName), "one") {
			costComponents = append(costComponents, &schema.CostComponent{
				Name:            "Storage (one zone, infrequent access)",
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: gbStorage,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(region),
					Service:       strPtr("AmazonEFS"),
					ProductFamily: strPtr("Storage"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr("/IATimedStorage-Z-ByteHrs/")},
					},
				},
			})
		}

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Read requests (infrequent access)",
			Unit:            "GB",
			UnitMultiplier:  1,
			MonthlyQuantity: infrequentAccessReadGbRequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEFS"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "accessType", Value: strPtr("Read")},
				},
			},
		})

		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Write requests (infrequent access)",
			Unit:            "GB",
			UnitMultiplier:  1,
			MonthlyQuantity: infrequentAccessWriteGbRequests,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("AmazonEFS"),
				ProductFamily: strPtr("Storage"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "accessType", Value: strPtr("Write")},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func calculateProvisionedThroughput(gbStorage *decimal.Decimal, throughput decimal.Decimal) *decimal.Decimal {
	if gbStorage == nil {
		gbStorage = &decimal.Zero
	}
	defaultThroughput := gbStorage.Mul(decimal.NewFromInt(730).Div(decimal.NewFromInt(20).Mul(decimal.NewFromInt(1))))
	totalProvisionedThroughput := throughput.Mul(decimal.NewFromInt(730))
	totalBillableProvisionedThroughput := totalProvisionedThroughput.Sub(defaultThroughput).Div(decimal.NewFromInt(730))

	if totalBillableProvisionedThroughput.IsPositive() {
		return &totalBillableProvisionedThroughput
	}

	return &decimal.Zero
}
