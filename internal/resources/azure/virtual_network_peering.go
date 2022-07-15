package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// VirtualNetworkPeering struct represents a VNET peering.
//

// Resource information: https://azure.microsoft.com/en-us/services/virtual-network/#overview
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/virtual-network/
type VirtualNetworkPeering struct {
	Address           string
	SourceRegion      string
	DestinationRegion string
	SourceZone        string
	DestinationZone   string

	MonthlyDataTransferGB *float64 `infracost_usage:"monthly_data_transfer_gb"`
}

var VirtualNetworkPeeringUsageSchema = []*schema.UsageItem{
	{Key: "monthly_data_transfer_gb", DefaultValue: 0, ValueType: schema.Float64},
}

func (r *VirtualNetworkPeering) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *VirtualNetworkPeering) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.ingressDataProcessedCostComponent(),
		r.egressDataProcessedCostComponent(),
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    VirtualNetworkPeeringUsageSchema,
		CostComponents: costComponents,
	}
}

func (r *VirtualNetworkPeering) egressDataProcessedCostComponent() *schema.CostComponent {
	if r.DestinationRegion == r.SourceRegion {
		return &schema.CostComponent{
			Name:            "Outbound data transfer",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataTransferGB),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr("Global"),
				Service:       strPtr("Virtual Network"),
				ProductFamily: strPtr("Networking"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "meterName", Value: strPtr("Intra-Region Egress")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		}
	}

	return &schema.CostComponent{
		Name:            "Outbound data transfer",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataTransferGB),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(r.SourceZone),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "serviceFamily", ValueRegex: regexPtr("Networking")},
				{Key: "productName", ValueRegex: regexPtr("VPN Gateway Bandwidth")},
				{Key: "meterName", ValueRegex: regexPtr("Inter-Virtual Network Data Transfer Out")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}

func (r *VirtualNetworkPeering) ingressDataProcessedCostComponent() *schema.CostComponent {
	if r.DestinationRegion == r.SourceRegion {
		return &schema.CostComponent{
			Name:            "Inbound data transfer",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataTransferGB),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("azure"),
				Region:        strPtr("Global"),
				Service:       strPtr("Virtual Network"),
				ProductFamily: strPtr("Networking"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "meterName", Value: strPtr("Intra-Region Ingress")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("Consumption"),
			},
		}
	}

	return &schema.CostComponent{
		Name:            "Inbound data transfer",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.MonthlyDataTransferGB),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("azure"),
			Region:     strPtr(r.DestinationZone),
			Service:    strPtr("VPN Gateway"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "serviceFamily", ValueRegex: regexPtr("Networking")},
				{Key: "productName", ValueRegex: regexPtr("VPN Gateway Bandwidth")},
				{Key: "meterName", ValueRegex: regexPtr("Inter-Virtual Network Data Transfer Out")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
