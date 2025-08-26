package aws

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

// GlobalacceleratorEndpointGroup struct represents a Global Accelerator endpoint group
//
// Resource information: https://docs.aws.amazon.com/global-accelerator/latest/dg/about-endpoint-groups.html
// Pricing information: https://aws.amazon.com/it/global-accelerator/pricing/
type GlobalacceleratorEndpointGroup struct {
	Address string
	Region  string

	MonthlyInboundDataTransferGB  *globalAcceleratorRegionDataTransferUsage `infracost_usage:"monthly_inbound_data_transfer_gb"`
	MonthlyOutboundDataTransferGB *globalAcceleratorRegionDataTransferUsage `infracost_usage:"monthly_outbound_data_transfer_gb"`
}

var globalAcceleratorRegionDataTransferUsageSchema = []*schema.UsageItem{
	{Key: "us", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "europe", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "south_africa", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "south_america", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "south_korea", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "australia", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "asia_pacific", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "india", DefaultValue: 0, ValueType: schema.Float64},
}
var regionCodeMapping = map[string]string{
	"us-gov-west-1":   "US",
	"us-gov-east-1":   "US",
	"us-east-1":       "NA",
	"us-east-2":       "NA",
	"us-west-1":       "NA",
	"us-west-2":       "NA",
	"us-west-2-lax-1": "NA",
	"ca-central-1":    "NA",
	"ca-west-1":       "NA",
	"mx-central-1":    "NA",
	"cn-north-1":      "AP",
	"cn-northwest-1":  "AP",
	"eu-central-1":    "EU",
	"eu-west-1":       "EU",
	"eu-west-2":       "EU",
	"eu-south-1":      "EU",
	"eu-west-3":       "EU",
	"eu-north-1":      "EU",
	"il-central-1":    "EU",
	"ap-east-1":       "AP",
	"ap-east-2":       "AP",
	"ap-northeast-1":  "AP",
	"ap-northeast-2":  "AP",
	"ap-northeast-3":  "AP",
	"ap-southeast-1":  "AP",
	"ap-southeast-2":  "AP",
	"ap-southeast-3":  "AP",
	"ap-southeast-4":  "AP",
	"ap-southeast-5":  "AP",
	"ap-southeast-7":  "AP",
	"ap-south-1":      "AP",
	"ap-south-2":      "AP",
	"me-central-1":    "ME",
	"me-south-1":      "ME",
	"sa-east-1":       "SA",
	"af-south-1":      "ZA",
}

type globalAcceleratorRegionDataTransferUsage struct {
	US           *float64 `infracost_usage:"us"`
	Europe       *float64 `infracost_usage:"europe"`
	SouthAfrica  *float64 `infracost_usage:"south_africa"`
	SouthAmerica *float64 `infracost_usage:"south_america"`
	SouthKorea   *float64 `infracost_usage:"south_korea"`
	Australia    *float64 `infracost_usage:"australia"`
	AsiaPacific  *float64 `infracost_usage:"asia_pacific"`
	MiddleEast   *float64 `infracost_usage:"middle_east"`
	India        *float64 `infracost_usage:"india"`
}

type globalAcceleratorRegionData struct {
	awsGroupedName                string
	codeRegion                    string
	monthlyInboundDataTransferGB  *float64
	monthlyOutboundDataTransferGB *float64
}

func (g *globalAcceleratorRegionData) HasUsage() bool {
	return g.monthlyInboundDataTransferGB != nil || g.monthlyOutboundDataTransferGB != nil
}

func (r *GlobalacceleratorEndpointGroup) CoreType() string {
	return "GlobalacceleratorEndpointGroup"
}

func (r *GlobalacceleratorEndpointGroup) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:          "monthly_inbound_data_transfer_gb",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_inbound_data_transfer_gb", Items: globalAcceleratorRegionDataTransferUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "monthly_outbound_data_transfer_gb",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_outbound_data_transfer_gb", Items: globalAcceleratorRegionDataTransferUsageSchema},
			ValueType:    schema.SubResourceUsage,
		},
	}
}

func (r *GlobalacceleratorEndpointGroup) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *GlobalacceleratorEndpointGroup) BuildResource() *schema.Resource {

	if r.MonthlyInboundDataTransferGB == nil {
		r.MonthlyInboundDataTransferGB = &globalAcceleratorRegionDataTransferUsage{}
	}
	if r.MonthlyOutboundDataTransferGB == nil {
		r.MonthlyOutboundDataTransferGB = &globalAcceleratorRegionDataTransferUsage{}
	}

	subResources := r.buildSubresources()

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: []*schema.CostComponent{},
		SubResources:   subResources,
	}
}

func (r *GlobalacceleratorEndpointGroup) buildSubresources() []*schema.Resource {
	var (
		totalInbound  float64 = 0
		totalOutbound float64 = 0
	)
	regionsData := []*globalAcceleratorRegionData{
		{
			awsGroupedName:                "US, Mexico, Canada",
			codeRegion:                    "NA",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.US,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.US,
		},
		{
			awsGroupedName:                "Europe",
			codeRegion:                    "EU",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.Europe,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.Europe,
		},
		{
			awsGroupedName:                "South Africa, Kenya",
			codeRegion:                    "ZA",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.SouthAfrica,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.SouthAfrica,
		},
		{
			awsGroupedName:                "South America",
			codeRegion:                    "SA",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.SouthAmerica,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.SouthAmerica,
		},
		{
			awsGroupedName:                "South Korea",
			codeRegion:                    "KR",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.SouthKorea,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.SouthKorea,
		},
		{
			awsGroupedName:                "Middle East",
			codeRegion:                    "ME",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.SouthKorea,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.SouthKorea,
		},
		{
			awsGroupedName:                "Australia, New Zealand",
			codeRegion:                    "AU",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.Australia,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.Australia,
		},
		{
			awsGroupedName:                "Asia Pacific",
			codeRegion:                    "AP",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.AsiaPacific,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.AsiaPacific,
		},
		{
			awsGroupedName:                "India, Indonesia, Philippines, Thailand",
			codeRegion:                    "IN",
			monthlyInboundDataTransferGB:  r.MonthlyInboundDataTransferGB.India,
			monthlyOutboundDataTransferGB: r.MonthlyOutboundDataTransferGB.India,
		},
	}
	trafficDirection := "In"

	for _, data := range regionsData {
		if !data.HasUsage() {
			continue
		}
		totalInbound += floatVal(data.monthlyInboundDataTransferGB)
		totalOutbound += floatVal(data.monthlyOutboundDataTransferGB)
	}
	if totalInbound < totalOutbound {
		trafficDirection = "Out"
	}

	subresources := []*schema.Resource{}

	for _, data := range regionsData {
		if !data.HasUsage() {
			continue
		}

		subresources = append(subresources, r.buildRegionSubresource(data, trafficDirection))
	}

	return subresources
}

func (r *GlobalacceleratorEndpointGroup) buildRegionSubresource(regionData *globalAcceleratorRegionData, trafficDirection string) *schema.Resource {
	from := regionCodeMapping[r.Region]
	to := regionData.codeRegion
	quantity := regionData.monthlyInboundDataTransferGB
	if trafficDirection == "Out" {
		quantity = regionData.monthlyOutboundDataTransferGB
	}

	// Even if there are multiple price record entries the price for -Bytes-Internet and -Bytes-AWS for the same regions are equal
	// So one of these two can be fixed to avoid multiple prices found
	usageType := fmt.Sprintf("%s-%s-%s-Bytes-Internet", strings.ToUpper(from), strings.ToUpper(to), strings.ToUpper(trafficDirection))
	resource := &schema.Resource{
		Name: regionData.awsGroupedName,
		CostComponents: []*schema.CostComponent{
			{
				Name:            fmt.Sprintf("%sbound DT-premium fee", trafficDirection),
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromFloat(floatVal(quantity))),
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Service:    strPtr("AWSGlobalAccelerator"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "trafficDirection", Value: strPtr(trafficDirection)},
						{Key: "fromLocation", Value: strPtr(from)},
						{Key: "toLocation", Value: strPtr(to)},
						{Key: "operation", Value: strPtr("Dominant")},
						{Key: "usagetype", Value: strPtr(usageType)},
					},
				},
				UsageBased: true,
			},
		},
	}

	return resource
}
