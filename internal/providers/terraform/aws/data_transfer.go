package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

var regionMapping = map[string]string{
	"us-gov-west-1":   "AWS GovCloud (US-West)",
	"us-gov-east-1":   "AWS GovCloud (US-East)",
	"us-east-1":       "US East (N. Virginia)",
	"us-east-2":       "US East (Ohio)",
	"us-west-1":       "US West (N. California)",
	"us-west-2":       "US West (Oregon)",
	"us-west-2-lax-1": "US West (Los Angeles)",
	"ca-central-1":    "Canada (Central)",
	"cn-north-1":      "China (Beijing)",
	"cn-northwest-1":  "China (Ningxia)",
	"eu-central-1":    "EU (Frankfurt)",
	"eu-west-1":       "EU (Ireland)",
	"eu-west-2":       "EU (London)",
	"eu-south-1":      "EU (Milan)",
	"eu-west-3":       "EU (Paris)",
	"eu-north-1":      "EU (Stockholm)",
	"ap-east-1":       "Asia Pacific (Hong Kong)",
	"ap-northeast-1":  "Asia Pacific (Tokyo)",
	"ap-northeast-2":  "Asia Pacific (Seoul)",
	"ap-northeast-3":  "Asia Pacific (Osaka)",
	"ap-southeast-1":  "Asia Pacific (Singapore)",
	"ap-southeast-2":  "Asia Pacific (Sydney)",
	"ap-south-1":      "Asia Pacific (Mumbai)",
	"me-south-1":      "Middle East (Bahrain)",
	"sa-east-1":       "South America (Sao Paulo)",
	"af-south-1":      "Africa (Cape Town)",
}

type dataTransferRegionUsageFilterData struct {
	usageNumber int64
	usageName   string
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
	region := u.Get("region").String()
	fromLocation, ok := regionMapping[region]

	if !ok {
		log.Warnf("Skipping resource %s. Could not find mapping for region %s", d.Address, region)
		return nil
	}

	usEastRegion := regionMapping["us-east-1"]
	otherRegion := regionMapping["us-west-1"]

	if region == "us-east-1" {
		usEastRegion = regionMapping["us-east-2"]
		otherRegion = regionMapping["us-west-2"]
	} else if region == "us-west-1" {
		otherRegion = regionMapping["us-west-2"]
	} else if region == "cn-north-1" {
		otherRegion = regionMapping["cn-northwest-1"]
	} else if region == "cn-northwest-1" {
		otherRegion = regionMapping["cn-north-1"]
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
			UnitMultiplier:  1,
			MonthlyQuantity: intraRegionGb,
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
			UnitMultiplier:  1,
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
			UnitMultiplier:  1,
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
	var used, lastEndUsageAmount int64
	for _, usageFilter := range usageFiltersData {
		usageName := usageFilter.usageName
		endUsageAmount := usageFilter.usageNumber
		var quantity *decimal.Decimal
		if endUsageAmount != 0 && usageAmount >= endUsageAmount {
			used = endUsageAmount - used
			lastEndUsageAmount = endUsageAmount
			quantity = decimalPtr(decimal.NewFromInt(used))
		} else if usageAmount > lastEndUsageAmount {
			used = usageAmount - lastEndUsageAmount
			lastEndUsageAmount = endUsageAmount
			quantity = decimalPtr(decimal.NewFromInt(used))
		}
		if quantity == nil {
			break
		}
		var usageFilter string
		if endUsageAmount != 0 {
			usageFilter = fmt.Sprint(endUsageAmount)
		} else {
			usageFilter = "Inf"
		}
		results = append(results, &UsageStepsFilterData{
			usageName:   usageName,
			usageFilter: usageFilter,
			quantity:    quantity,
		})
	}
	return results
}

func outboundInternet(fromLocation string, networkUsage int64) []*schema.CostComponent {
	costComponents := make([]*schema.CostComponent, 0)
	defaultUsageFiltersData := []*dataTransferRegionUsageFilterData{
		{
			usageName:   "first 10TB",
			usageNumber: 10240,
		},
		{
			usageName:   "next 40TB",
			usageNumber: 51200,
		},
		{
			usageName:   "next 100TB",
			usageNumber: 153600,
		},
		{
			usageName:   "over 150TB",
			usageNumber: 0,
		},
	}
	chinaUsageFiltersData := []*dataTransferRegionUsageFilterData{
		{
			usageName:   "",
			usageNumber: 0,
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
			UnitMultiplier:  1,
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
