package aws

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

// DataTransfer represents data transferred "in" to and "out" of Amazon EC2.
//
// Pricing information here: https://aws.amazon.com/ec2/pricing/on-demand/
type DataTransfer struct {
	Address string
	Region  string

	// "usage" args
	MonthlyInfraRegionGB            *float64 `infracost_usage:"monthly_intra_region_gb"`
	MonthlyOutboundInternetGB       *float64 `infracost_usage:"monthly_outbound_internet_gb"`
	MonthlyOutboundUsEastToUsEastGB *float64 `infracost_usage:"monthly_outbound_us_east_to_us_east_gb"`
	MonthlyOutboundOtherRegionsGB   *float64 `infracost_usage:"monthly_outbound_other_regions_gb"`
}

func (r *DataTransfer) CoreType() string {
	return "DataTransfer"
}

func (r *DataTransfer) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_intra_region_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_outbound_internet_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_outbound_us_east_to_us_east_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_outbound_other_regions_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the DataTransfer.
// It uses the `infracost_usage` struct tags to populate data into the DataTransfer.
func (r *DataTransfer) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid DataTransfer.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *DataTransfer) BuildResource() *schema.Resource {
	_, ok := RegionMapping[r.Region]

	if !ok {
		logging.Logger.Warn().Msgf("Skipping resource %s. Could not find mapping for region %s", r.Address, r.Region)
		return nil
	}

	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, r.intraRegionCostComponents()...)
	costComponents = append(costComponents, r.outboundInternetCostComponents()...)
	costComponents = append(costComponents, r.outboundUsEastCostComponents()...)
	costComponents = append(costComponents, r.outboundOtherRegionsCostComponents()...)

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
	}
}

// intraRegionCostComponents returns a cost component for Intra-region data
// transfer only when its usage is specified.
func (r *DataTransfer) intraRegionCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.MonthlyInfraRegionGB == nil {
		return costComponents
	}

	intraRegionGb := decimalPtr(decimal.NewFromFloat(*r.MonthlyInfraRegionGB))

	if intraRegionGb != nil {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Intra-region data transfer",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(intraRegionGb.Mul(decimal.NewFromInt(2))),
			ProductFilter:   r.buildProductFilter("IntraRegion", nil, "DataTransfer-Regional-Bytes"),
			UsageBased:      true,
		})
	}

	return costComponents
}

// intraRegionCostComponents returns a cost component for outbound data
// transfer to the Internet only when its usage is specified.
// China regions are calculated without tiers.
func (r *DataTransfer) outboundInternetCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.MonthlyOutboundInternetGB == nil {
		return costComponents
	}

	outboundInternetGb := decimalPtr(decimal.NewFromFloat(*r.MonthlyOutboundInternetGB))
	networkUsage := outboundInternetGb.IntPart()

	if r.Region == "cn-north-1" || r.Region == "cn-northwest-1" {
		costComponents = append(costComponents, r.buildOutboundInternetCostComponent(
			"Outbound data transfer to Internet",
			decimal.NewFromInt(networkUsage),
			"Inf",
		))
		return costComponents
	}

	type dataTransferRegionUsageFilterData struct {
		usageName      string
		tierCapacity   int64
		endUsageNumber int64
	}

	usageFiltersData := []*dataTransferRegionUsageFilterData{
		{
			usageName:      "first 10TB",
			tierCapacity:   10240,
			endUsageNumber: 10240,
		},
		{
			usageName:      "next 40TB",
			tierCapacity:   40960,
			endUsageNumber: 51200,
		},
		{
			usageName:      "next 100TB",
			tierCapacity:   102400,
			endUsageNumber: 153600,
		},
		{
			usageName:      "over 150TB",
			tierCapacity:   0,
			endUsageNumber: 0,
		},
	}

	tierLimits := make([]int, len(usageFiltersData)-1)
	for i, usageFilter := range usageFiltersData {
		if usageFilter.tierCapacity > 0 {
			tierLimits[i] = int(usageFilter.tierCapacity)
		}
	}

	tiers := usage.CalculateTierBuckets(decimal.NewFromInt(networkUsage), tierLimits)
	for i, usageFilter := range usageFiltersData {
		quantity := tiers[i]

		if quantity.Equals(decimal.Zero) {
			break
		}

		filter := fmt.Sprint(usageFilter.endUsageNumber)
		if usageFilter.endUsageNumber == 0 {
			filter = "Inf"
		}

		name := fmt.Sprintf("Outbound data transfer to Internet (%s)", usageFilter.usageName)
		costComponents = append(costComponents, r.buildOutboundInternetCostComponent(
			name,
			quantity,
			filter,
		))
	}

	return costComponents
}

// buildOutboundInternetCostComponent builds a cost component for
// outbound data transfer to Internet.
func (r *DataTransfer) buildOutboundInternetCostComponent(name string, networkUsage decimal.Decimal, endUsageAmount string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            name,
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(networkUsage),
		ProductFilter:   r.buildProductFilter("AWS Outbound", nil, ""),
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(endUsageAmount),
		},
		UsageBased: true,
	}
}

// outboundUsEastCostComponents returns a cost component for outbound data
// transfer to US East regions only when its usage is specified.
func (r *DataTransfer) outboundUsEastCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.MonthlyOutboundUsEastToUsEastGB == nil {
		return costComponents
	}

	toRegion := "us-east-1"

	if r.Region == "us-east-1" {
		toRegion = "us-east-2"
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Outbound data transfer to US East regions",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(*r.MonthlyOutboundUsEastToUsEastGB)),
		ProductFilter:   r.buildProductFilter("InterRegion Outbound", &toRegion, ""),
		UsageBased:      true,
	})

	return costComponents
}

// outboundOtherRegionsCostComponents returns a cost component for outbound data
// transfer to other regions only when its usage is specified.
func (r *DataTransfer) outboundOtherRegionsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.MonthlyOutboundOtherRegionsGB == nil {
		return costComponents
	}

	toRegion := "us-west-1"

	switch r.Region {
	case "us-east-1":
		toRegion = "us-west-2"
	case "us-west-1":
		toRegion = "us-west-2"
	case "cn-north-1":
		toRegion = "cn-northwest-1"
	case "cn-northwest-1":
		toRegion = "cn-north-1"
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Outbound data transfer to other regions",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(*r.MonthlyOutboundOtherRegionsGB)),
		ProductFilter:   r.buildProductFilter("InterRegion Outbound", &toRegion, ""),
		UsageBased:      true,
	})

	return costComponents
}

// buildProductFilter returns a filter for data transfer products. Desctination
// region is optional.
func (r *DataTransfer) buildProductFilter(transferType string, toRegion *string, usageTypeSuffix string) *schema.ProductFilter {
	fromLocation := RegionMapping[r.Region]

	attributeFilters := []*schema.AttributeFilter{
		{Key: "transferType", Value: strPtr(transferType)},
		{Key: "fromLocation", Value: strPtr(fromLocation)},
	}

	if toRegion != nil {
		toLocation := RegionMapping[*toRegion]
		attributeFilters = append(attributeFilters, &schema.AttributeFilter{
			Key:   "toLocation",
			Value: strPtr(toLocation),
		})
	}

	if regionCode, ok := RegionCodeMapping[r.Region]; ok {
		attributeFilters = append(attributeFilters, &schema.AttributeFilter{
			Key:        "usagetype",
			ValueRegex: regexPtr(fmt.Sprintf("^%s-", regionCode)),
		})
	}

	if usageTypeSuffix != "" {
		attributeFilters = append(attributeFilters, &schema.AttributeFilter{
			Key:        "usagetype",
			ValueRegex: regexPtr(fmt.Sprintf("%s$", usageTypeSuffix)),
		})
	}

	return &schema.ProductFilter{
		VendorName:       strPtr("aws"),
		Service:          strPtr("AWSDataTransfer"),
		ProductFamily:    strPtr("Data Transfer"),
		AttributeFilters: attributeFilters,
	}
}
