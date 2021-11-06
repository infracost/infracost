package aws

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// TransferServer defines a AWS Transfer Server resource from Transfer Family
// service. It supports multiple transfer protocols like FTP/FTPS/SFTP and
// each is billed hourly when enabled. It also bills the amount of data
// being downloaded/uploaded over those protocols.
//
// See more resource information here: https://aws.amazon.com/aws-transfer-family/.
//
// See the pricing information here: https://aws.amazon.com/aws-transfer-family/pricing/.
type TransferServer struct {
	Address   string
	Region    string
	Protocols []string

	// "usage" args
	MonthlyDataDownloadedGB *float64 `infracost_usage:"monthly_data_downloaded_gb"`
	MonthlyDataUploadedGB   *float64 `infracost_usage:"monthly_data_uploaded_gb"`
}

// TransferServerUsageSchema defines a list of usage items for TransferServer.
var TransferServerUsageSchema = []*schema.UsageItem{
	{Key: "monthly_data_downloaded_gb", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "monthly_data_uploaded_gb", DefaultValue: 0, ValueType: schema.Float64},
}

// PopulateUsage parses the u schema.UsageData into the TransferServer.
// It uses the `infracost_usage` struct tags to populate data into the TransferServer.
func (t *TransferServer) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(t, u)
}

// BuildResource builds a schema.Resource from a valid TransferServer struct.
// This method is called after the resource is initialised by an IaC provider.
func (t *TransferServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	service := "AWSTransfer"
	productFamily := "AWS Transfer Family"

	for _, protocol := range t.Protocols {
		costComponents = append(costComponents,
			newProtocolCostComponent(
				protocol,
				"[A-Z0-9]*-ProtocolHours",
				t.Region,
				service,
				productFamily,
			),
		)
	}

	costComponents = append(costComponents,
		newDataTransferCostComponent(
			"Data downloaded",
			t.MonthlyDataDownloadedGB,
			"[A-Z0-9]*-DownloadBytes",
			t.Region,
			service,
			productFamily,
		),
		newDataTransferCostComponent(
			"Data uploaded",
			t.MonthlyDataUploadedGB,
			"[A-Z0-9]*-UploadBytes",
			t.Region,
			service,
			productFamily,
		),
	)

	return &schema.Resource{
		Name:           t.Address,
		UsageSchema:    TransferServerUsageSchema,
		CostComponents: costComponents,
	}
}

func newProtocolCostComponent(protocol string, usageType string, region string, service string, productFamily string) *schema.CostComponent {
	// This value can be used for any storage type as their pricing is
	// identical, but for some protocols EFS prices are missing in the pricing API.
	storageType := "S3"

	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s protocol enabled", protocol),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr(service),
			ProductFamily: strPtr(productFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", usageType))},
				{Key: "operation", ValueRegex: strPtr(fmt.Sprintf("/^%s:%s$/i", protocol, storageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func newDataTransferCostComponent(name string, quantity *float64, usageType string, region string, service string, productFamily string) *schema.CostComponent {
	// This value can be used for any protocol/storage type as their pricing is
	// identical, but for some protocols EFS prices are missing in the pricing API.
	storageType := "FTP:S3"

	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(quantity),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(region),
			Service:       strPtr(service),
			ProductFamily: strPtr(productFamily),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", usageType))},
				{Key: "operation", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", storageType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}
