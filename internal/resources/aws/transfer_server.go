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
func (r *TransferServer) CoreType() string {
	return "TransferServer"
}

func (r *TransferServer) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_data_downloaded_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_data_uploaded_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the TransferServer.
// It uses the `infracost_usage` struct tags to populate data into the TransferServer.
func (r *TransferServer) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid TransferServer struct.
// This method is called after the resource is initialised by an IaC provider.
func (r *TransferServer) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	for _, protocol := range r.Protocols {
		costComponents = append(costComponents, r.protocolEnabledCostComponent(protocol))
	}

	costComponents = append(costComponents, r.dataDownloadedCostComponent())
	costComponents = append(costComponents, r.dataUploadedCostComponent())

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *TransferServer) protocolEnabledCostComponent(protocol string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s protocol enabled", protocol),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter:  r.buildProductFilter(protocol, "^[A-Z0-9]*-ProtocolHours$"),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
	}
}

func (r *TransferServer) dataDownloadedCostComponent() *schema.CostComponent {
	// The pricing is identical for all protocols and the traffic is combined
	transferProtocol := "FTP"

	return &schema.CostComponent{
		Name:            "Data downloaded",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataDownloadedGB),
		ProductFilter:   r.buildProductFilter(transferProtocol, "^[A-Z0-9]*-DownloadBytes$"),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *TransferServer) dataUploadedCostComponent() *schema.CostComponent {
	// The pricing is identical for all protocols and the traffic is combined
	transferProtocol := "FTP"

	return &schema.CostComponent{
		Name:            "Data uploaded",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataUploadedGB),
		ProductFilter:   r.buildProductFilter(transferProtocol, "^[A-Z0-9]*-UploadBytes$"),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
}

func (r *TransferServer) buildProductFilter(protocol, usageType string) *schema.ProductFilter {
	// The pricing for all storage types is identical, but for some protocols
	// EFS prices are missing in the pricing API.
	storageType := "S3"

	return &schema.ProductFilter{
		VendorName:    strPtr("aws"),
		Region:        strPtr(r.Region),
		Service:       strPtr("AWSTransfer"),
		ProductFamily: strPtr("AWS Transfer Family"),
		AttributeFilters: []*schema.AttributeFilter{
			{Key: "usagetype", ValueRegex: regexPtr(usageType)},
			{Key: "operation", ValueRegex: regexPtr(fmt.Sprintf("^%s:%s$", protocol, storageType))},
		},
	}
}
