package aws

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

type dataTransferRegionUsageFilterData struct {
	usageName      string
	tierCapacity   int64
	endUsageNumber int64
}

type UsageStepsFilterData struct {
	usageName   string
	usageFilter string
	quantity    *decimal.Decimal
}

func GetDataTransferRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_data_transfer",
		RFunc: NewDataTransfer,
	}
}

func NewDataTransfer(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := strings.ToLower(u.Get("region").String())
	fromLocation, ok := aws.RegionMapping[region]

	if !ok {
		log.Warnf("Skipping resource %s. Could not find mapping for region %s", d.Address, region)
		return nil
	}

	usEastRegion := aws.RegionMapping["us-east-1"]
	otherRegion := aws.RegionMapping["us-west-1"]

	if region == "us-east-1" {
		usEastRegion = aws.RegionMapping["us-east-2"]
		otherRegion = aws.RegionMapping["us-west-2"]
	} else if region == "us-west-1" {
		otherRegion = aws.RegionMapping["us-west-2"]
	} else if region == "cn-north-1" {
		otherRegion = aws.RegionMapping["cn-northwest-1"]
	} else if region == "cn-northwest-1" {
		otherRegion = aws.RegionMapping["cn-north-1"]
	}

	var intraRegionGb *decimal.Decimal
	if u != nil && u.Get("monthly_intra_region_gb").Exists() {
		intraRegionGb = decimalPtr(decimal.NewFromFloat(u.Get("monthly_intra_region_gb").Float()))
	}

	var outboundInternetGb *decimal.Decimal
	if u != nil && u.Get("monthly_outbound_internet_gb").Exists() {
		outboundInternetGb = decimalPtr(decimal.NewFromFloat(u.Get("monthly_outbound_internet_gb").Float()))
	}

	var outboundUsEastGb *decimal.Decimal
	if u != nil && u.Get("monthly_outbound_us_east_to_us_east_gb").Exists() {
		outboundUsEastGb = decimalPtr(decimal.NewFromFloat(u.Get("monthly_outbound_us_east_to_us_east_gb").Float()))
	}

	var outboundOtherRegionsGb *decimal.Decimal
	if u != nil && u.Get("monthly_outbound_other_regions_gb").Exists() {
		outboundOtherRegionsGb = decimalPtr(decimal.NewFromFloat(u.Get("monthly_outbound_other_regions_gb").Float()))
	}

	costComponents := make([]*schema.CostComponent, 0)

	if intraRegionGb != nil {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Intra-region data transfer",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(intraRegionGb.Mul(decimal.NewFromInt(2))),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Service:       strPtr("AWSDataTransfer"),
				ProductFamily: strPtr("Data Transfer"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "transferType", Value: strPtr("IntraRegion")},
					{Key: "fromLocation", Value: strPtr(fromLocation)},
				},
			},
		})
	}

	if outboundInternetGb != nil {
		costComponents = append(costComponents, outboundInternet(fromLocation, outboundInternetGb.IntPart())...)
	}

	if outboundUsEastGb != nil {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Outbound data transfer to US East regions",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: outboundUsEastGb,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Service:       strPtr("AWSDataTransfer"),
				ProductFamily: strPtr("Data Transfer"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "transferType", Value: strPtr("InterRegion Outbound")},
					{Key: "fromLocation", Value: strPtr(fromLocation)},
					{Key: "toLocation", Value: strPtr(usEastRegion)},
				},
			},
		})
	}

	if outboundOtherRegionsGb != nil {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Outbound data transfer to other regions",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: outboundOtherRegionsGb,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Service:       strPtr("AWSDataTransfer"),
				ProductFamily: strPtr("Data Transfer"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "transferType", Value: strPtr("InterRegion Outbound")},
					{Key: "fromLocation", Value: strPtr(fromLocation)},
					{Key: "toLocation", Value: strPtr(otherRegion)},
				},
			},
		})
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func usageStepsFilterHelper(usageFiltersData []*dataTransferRegionUsageFilterData, usageAmount int64) []*UsageStepsFilterData {
	results := make([]*UsageStepsFilterData, 0)
	if len(usageFiltersData) == 1 {
		quantity := decimal.NewFromInt(usageAmount)
		results = append(results, &UsageStepsFilterData{
			usageName:   usageFiltersData[0].usageName,
			usageFilter: "Inf",
			quantity:    &quantity,
		})
		return results
	}
	tierLimits := make([]int, len(usageFiltersData)-1)
	for idx, usageFilter := range usageFiltersData {
		if usageFilter.tierCapacity > 0 {
			tierLimits[idx] = int(usageFilter.tierCapacity)
		}
	}
	tiers := usage.CalculateTierBuckets(decimal.NewFromInt(usageAmount), tierLimits)
	for idx, usageFilter := range usageFiltersData {
		if tiers[idx].Equals(decimal.Zero) {
			break
		}
		filter := fmt.Sprint(usageFilter.endUsageNumber)
		if usageFilter.endUsageNumber == 0 {
			filter = "Inf"
		}
		results = append(results, &UsageStepsFilterData{
			usageName:   usageFilter.usageName,
			usageFilter: filter,
			quantity:    &tiers[idx],
		})
	}
	return results
}

func outboundInternet(fromLocation string, networkUsage int64) []*schema.CostComponent {
	costComponents := make([]*schema.CostComponent, 0)
	defaultUsageFiltersData := []*dataTransferRegionUsageFilterData{
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
	chinaUsageFiltersData := []*dataTransferRegionUsageFilterData{
		{
			usageName:      "",
			tierCapacity:   0,
			endUsageNumber: 0,
		},
	}
	usageFiltersData := defaultUsageFiltersData
	chinaLocations := map[string]struct{}{"China (Beijing)": {}, "China (Ningxia)": {}}
	if _, ok := chinaLocations[fromLocation]; ok {
		usageFiltersData = chinaUsageFiltersData
	}
	for _, usageStep := range usageStepsFilterHelper(usageFiltersData, networkUsage) {
		var name string
		if usageStep.usageName == "" {
			name = "Outbound data transfer to Internet"
		} else {
			name = fmt.Sprintf("Outbound data transfer to Internet (%s)", usageStep.usageName)
		}
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            name,
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: usageStep.quantity,
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Service:       strPtr("AWSDataTransfer"),
				ProductFamily: strPtr("Data Transfer"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "transferType", Value: strPtr("AWS Outbound")},
					{Key: "fromLocation", Value: strPtr(fromLocation)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr(usageStep.usageFilter),
			},
		})
	}
	return costComponents
}
