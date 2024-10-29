package aws

import (
	"fmt"
	"sort"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"

	"github.com/shopspring/decimal"
)

type DXConnection struct {
	Address                                         string
	Bandwidth                                       string
	Location                                        string
	Region                                          string
	MonthlyOutboundFromRegionToDXConnectionLocation *RegionsUsage `infracost_usage:"monthly_outbound_from_region_to_dx_connection_location"`
	MonthlyOutboundRegionToDxLocationGB             *float64      `infracost_usage:"monthly_outbound_region_to_dx_location_gb"`
	DxVirtualInterfaceType                          *string       `infracost_usage:"dx_virtual_interface_type"`
	DXConnectionType                                *string       `infracost_usage:"dx_connection_type"`
}

func (r *DXConnection) CoreType() string {
	return "DXConnection"
}

func (r *DXConnection) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:          "monthly_outbound_from_region_to_dx_connection_location",
			ValueType:    schema.SubResourceUsage,
			DefaultValue: &usage.ResourceUsage{Name: "monthly_outbound_from_region_to_dx_connection_location", Items: RegionUsageSchema},
		},
		{Key: "monthly_outbound_region_to_dx_location_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "dx_virtual_interface_type", ValueType: schema.String, DefaultValue: "private"},
		{Key: "dx_connection_type", ValueType: schema.String, DefaultValue: "dedicated"},
	}
}

func (r *DXConnection) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DXConnection) BuildResource() *schema.Resource {
	components := []*schema.CostComponent{r.connectionComponent()}

	if r.MonthlyOutboundFromRegionToDXConnectionLocation != nil {
		regionUsages := r.MonthlyOutboundFromRegionToDXConnectionLocation.Values()
		sort.Slice(regionUsages, func(i, j int) bool {
			return regionUsages[i].Key < regionUsages[j].Key
		})

		for _, regionUsage := range regionUsages {
			dataProcessedGB := decimalPtr(decimal.NewFromFloat(regionUsage.Value))
			outboundDataTransferComponent := r.outboundDataTransferComponent(regionUsage.Key, dataProcessedGB)

			if outboundDataTransferComponent != nil {
				components = append(components, outboundDataTransferComponent)
			}
		}
	}

	if r.MonthlyOutboundFromRegionToDXConnectionLocation == nil && r.MonthlyOutboundRegionToDxLocationGB != nil {
		dataProcessedGB := decimalPtr(decimal.NewFromFloat(*r.MonthlyOutboundRegionToDxLocationGB))
		outboundDataTransferComponent := r.outboundDataTransferComponent(r.Region, dataProcessedGB)

		if outboundDataTransferComponent != nil {
			components = append(components, outboundDataTransferComponent)
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: components,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *DXConnection) connectionComponent() *schema.CostComponent {
	bandwidth := strings.Replace(r.Bandwidth, "bps", "", 1)

	connectionType := "dedicated"
	if r.DXConnectionType != nil {
		connectionType = *r.DXConnectionType
	}

	return &schema.CostComponent{
		Name:           "DX connection",
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AWSDirectConnect"),
			ProductFamily: strPtr("Direct Connect"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "capacity", Value: strPtr(bandwidth)},
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", r.Location))},
				{Key: "connectionType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", connectionType))},
			},
		},
	}
}

func (r *DXConnection) outboundDataTransferComponent(fromRegion string, dataProcessedGB *decimal.Decimal) *schema.CostComponent {
	virtualInterfaceType := "private"
	if r.DxVirtualInterfaceType != nil {
		virtualInterfaceType = *r.DxVirtualInterfaceType
	}

	fromLocation, ok := RegionMapping[fromRegion]
	if !ok {
		// This shouldn't happen because we're loading the regions into the RegionsUsage struct
		// which should have same keys as the RegionMappings map
		logging.Logger.Warn().Msgf("Skipping resource %s usage cost: Outbound data transfer. Could not find mapping for region %s", r.Address, fromRegion)
		return nil
	}

	return &schema.CostComponent{
		Name:            fmt.Sprintf("Outbound data transfer (from %s, to %s)", fromRegion, r.Location),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: dataProcessedGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Service:       strPtr("AWSDirectConnect"),
			ProductFamily: strPtr("Data Transfer"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "fromLocation", Value: strPtr(fromLocation)},
				{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s-DataXfer-Out/", r.Location))},
				{Key: "virtualInterfaceType", ValueRegex: strPtr(fmt.Sprintf("/%s/i", virtualInterfaceType))},
			},
		},
		UsageBased: true,
	}
}
