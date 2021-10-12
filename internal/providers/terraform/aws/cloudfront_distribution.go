package aws

import (
	"fmt"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"strconv"
)

type regionData struct {
	awsGroupedName string
	priceRegion    string
	usageKey       string
}

func GetCloudfrontDistributionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudfront_distribution",
		RFunc: NewCloudfrontDistribution,
	}
}

func NewCloudfrontDistribution(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			encryptionRequests(u),
			realtimeLogs(u),
			customSSLCertificate(u),
		},
		SubResources: []*schema.Resource{
			regionalDataOutToInternet(u),
			regionalDataOutToOrigin(u),
			httpRequests(u),
			httpsRequests(u),
			shieldRequests(u),
		},
	}
	resource.CostComponents = append(resource.CostComponents, invalidationRequests(u)...)
	return resource
}

func regionalDataOutToInternet(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:         "Data transfer out to internet",
		SubResources: []*schema.Resource{},
	}

	var uMap map[string]gjson.Result
	if u != nil && u.Get("monthly_data_transfer_to_internet_gb").Exists() {
		uMap = u.Get("monthly_data_transfer_to_internet_gb").Map()
	}

	regionsData := []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "us",
		},
		{
			awsGroupedName: "Europe, Israel",
			priceRegion:    "Europe",
			usageKey:       "europe",
		},
		{
			awsGroupedName: "South Africa, Kenya, Middle East",
			priceRegion:    "South Africa",
			usageKey:       "south_africa",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "japan",
		},
		{
			awsGroupedName: "Australia, New Zealand",
			priceRegion:    "Australia",
			usageKey:       "australia",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "asia_pacific",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "india",
		},
	}

	tierStarts := []int{0, 10240, 51200, 153600, 512000, 1048576, 5242880}
	tierLimits := []int{10240, 40960, 102400, 358400, 536576, 4194304}
	tierNames := []string{"first 10TB", "next 40TB", "next 100TB", "next 350TB", "next 524TB", "next 4PB", "over 5PB"}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		fromLocation := regData.priceRegion
		usageKey := regData.usageKey

		var quantity *decimal.Decimal
		if _, ok := uMap[usageKey]; ok {
			quantity = decimalPtr(decimal.NewFromInt(uMap[usageKey].Int()))
		}

		if quantity == nil {
			resource.CostComponents = append(resource.CostComponents,
				dataOutCostComponent(awsRegion, tierNames[0], fromLocation, 0, nil))
			continue
		}

		tiers := usage.CalculateTierBuckets(*quantity, tierLimits)
		for i := range tiers {
			if tiers[i].GreaterThan(decimal.Zero) {
				resource.CostComponents = append(resource.CostComponents,
					dataOutCostComponent(awsRegion, tierNames[i], fromLocation, tierStarts[i], &tiers[i]))
			}
		}
	}

	return resource
}

func dataOutCostComponent(awsRegion, usageName, fromLocation string, startUsage int, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:            fmt.Sprintf("%v (%v)", awsRegion, usageName),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AmazonCloudFront"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "transferType", Value: strPtr("CloudFront Outbound")},
				{Key: "fromLocation", Value: strPtr(fromLocation)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr(strconv.Itoa(startUsage)),
		},
	}
}

func regionalDataOutToOrigin(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:           "Data transfer out to origin",
		CostComponents: []*schema.CostComponent{},
	}

	var uMap map[string]gjson.Result
	if u != nil && u.Get("monthly_data_transfer_to_origin_gb").Exists() {
		uMap = u.Get("monthly_data_transfer_to_origin_gb").Map()
	}

	regionsData := []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "us",
		},
		{
			awsGroupedName: "Europe, Israel",
			priceRegion:    "Europe",
			usageKey:       "europe",
		},
		{
			awsGroupedName: "South Africa, Kenya, Middle East",
			priceRegion:    "South Africa",
			usageKey:       "south_africa",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "japan",
		},
		{
			awsGroupedName: "Australia, New Zealand",
			priceRegion:    "Australia",
			usageKey:       "australia",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "asia_pacific",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "india",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey
		var quantity *decimal.Decimal
		if _, ok := uMap[usageKey]; ok {
			quantity = decimalPtr(decimal.NewFromInt(uMap[usageKey].Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: quantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Service:    strPtr("AmazonCloudFront"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "transferType", Value: strPtr("CloudFront to Origin")},
					{Key: "fromLocation", Value: strPtr(apiRegion)},
				},
			},
		})
	}

	return resource
}

func httpRequests(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:         "HTTP requests",
		SubResources: []*schema.Resource{},
	}

	var uMap map[string]gjson.Result
	if u != nil && u.Get("monthly_https_requests").Exists() {
		uMap = u.Get("monthly_https_requests").Map()
	}

	regionsData := []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "us",
		},
		{
			awsGroupedName: "Europe, Israel",
			priceRegion:    "Europe",
			usageKey:       "europe",
		},
		{
			awsGroupedName: "South Africa, Kenya, Middle East",
			priceRegion:    "South Africa",
			usageKey:       "south_africa",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "japan",
		},
		{
			awsGroupedName: "Australia, New Zealand",
			priceRegion:    "Australia",
			usageKey:       "australia",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "asia_pacific",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "india",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey
		var quantity *decimal.Decimal
		if _, ok := uMap[usageKey]; ok {
			quantity = decimalPtr(decimal.NewFromInt(uMap[usageKey].Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "10k requests",
			UnitMultiplier:  decimal.NewFromInt(10000),
			MonthlyQuantity: quantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Service:    strPtr("AmazonCloudFront"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "location", Value: strPtr(apiRegion)},
					{Key: "requestType", Value: strPtr("CloudFront-Request-HTTP-Proxy")},
				},
			},
		})
	}

	return resource
}

func httpsRequests(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:         "HTTPS requests",
		SubResources: []*schema.Resource{},
	}

	var uMap map[string]gjson.Result
	if u != nil && u.Get("monthly_https_requests").Exists() {
		uMap = u.Get("monthly_https_requests").Map()
	}

	regionsData := []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "us",
		},
		{
			awsGroupedName: "Europe, Israel",
			priceRegion:    "Europe",
			usageKey:       "europe",
		},
		{
			awsGroupedName: "South Africa, Kenya, Middle East",
			priceRegion:    "South Africa",
			usageKey:       "south_africa",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "japan",
		},
		{
			awsGroupedName: "Australia, New Zealand",
			priceRegion:    "Australia",
			usageKey:       "australia",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "asia_pacific",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "india",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey
		var quantity *decimal.Decimal
		if _, ok := uMap[usageKey]; ok {
			quantity = decimalPtr(decimal.NewFromInt(uMap[usageKey].Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "10k requests",
			UnitMultiplier:  decimal.NewFromInt(10000),
			MonthlyQuantity: quantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Service:    strPtr("AmazonCloudFront"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "location", Value: strPtr(apiRegion)},
					{Key: "requestType", Value: strPtr("CloudFront-Request-HTTPS-Proxy")},
				},
			},
		})
	}

	return resource
}

func shieldRequests(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:           "Origin shield HTTP requests",
		CostComponents: []*schema.CostComponent{},
	}

	var uMap map[string]gjson.Result
	if u != nil && u.Get("monthly_shield_requests").Exists() {
		uMap = u.Get("monthly_shield_requests").Map()
	}

	regionsData := []*regionData{
		{
			awsGroupedName: "US",
			priceRegion:    "US East (N. Virginia)",
			usageKey:       "us",
		},
		{
			awsGroupedName: "Europe",
			priceRegion:    "EU (Frankfurt)",
			usageKey:       "europe",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America (Sao Paulo)",
			usageKey:       "south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Asia Pacific (Tokyo)",
			usageKey:       "japan",
		},
		{
			awsGroupedName: "Australia",
			priceRegion:    "Asia Pacific (Sydney)",
			usageKey:       "australia",
		},
		{
			awsGroupedName: "Singapore",
			priceRegion:    "Asia Pacific (Singapore)",
			usageKey:       "singapore",
		},
		{
			awsGroupedName: "South Korea",
			priceRegion:    "Asia Pacific (Seoul)",
			usageKey:       "south_korea",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "Asia Pacific (Mumbai)",
			usageKey:       "india",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey
		var quantity *decimal.Decimal
		if _, ok := uMap[usageKey]; ok {
			quantity = decimalPtr(decimal.NewFromInt(uMap[usageKey].Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "10k requests",
			UnitMultiplier:  decimal.NewFromInt(10000),
			MonthlyQuantity: quantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Service:    strPtr("AmazonCloudFront"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "requestDescription", Value: strPtr("Origin Shield Requests")},
					{Key: "location", Value: strPtr(apiRegion)},
				},
			},
		})
	}

	return resource
}

func invalidationRequests(u *schema.UsageData) []*schema.CostComponent {
	var freeQuantity *decimal.Decimal
	var paidQuantity *decimal.Decimal
	if u != nil && u.Get("monthly_invalidation_requests").Exists() {
		usageAmount := u.Get("monthly_invalidation_requests").Int()
		if usageAmount < 1000 {
			freeQuantity = decimalPtr(decimal.NewFromInt(usageAmount))
		} else {
			freeQuantity = decimalPtr(decimal.NewFromInt(1000))
			paidQuantity = decimalPtr(decimal.NewFromInt(usageAmount - 1000))
		}
	}

	costComponents := []*schema.CostComponent{
		{
			Name:            "Invalidation requests (first 1k)",
			Unit:            "paths",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: freeQuantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Service:    strPtr("AmazonCloudFront"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("Invalidations")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("0"),
			},
		},
	}

	if paidQuantity != nil {
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            "Invalidation requests (over 1k)",
			Unit:            "paths",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: paidQuantity,
			ProductFilter: &schema.ProductFilter{
				VendorName: strPtr("aws"),
				Service:    strPtr("AmazonCloudFront"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "usagetype", Value: strPtr("Invalidations")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("1000"),
			},
		})
	}

	return costComponents
}

func encryptionRequests(u *schema.UsageData) *schema.CostComponent {
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_encryption_requests").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_encryption_requests").Int()))
	}
	return &schema.CostComponent{
		Name:            "Field level encryption requests",
		Unit:            "10k requests",
		UnitMultiplier:  decimal.NewFromInt(10000),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AmazonCloudFront"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "requestDescription", Value: strPtr("HTTPS Proxy requests with Field Level Encryption")},
				{Key: "location", Value: strPtr("Europe")},
			},
		},
	}
}

func realtimeLogs(u *schema.UsageData) *schema.CostComponent {
	var quantity *decimal.Decimal
	if u != nil && u.Get("monthly_log_lines").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_log_lines").Int()))
	}
	return &schema.CostComponent{
		Name:            "Real-time log requests",
		Unit:            "1M lines",
		UnitMultiplier:  decimal.NewFromInt(1000000),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AmazonCloudFront"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "operation", Value: strPtr("RealTimeLog")},
			},
		},
	}
}

func customSSLCertificate(u *schema.UsageData) *schema.CostComponent {
	var quantity *decimal.Decimal
	if u != nil && u.Get("custom_ssl_certificates").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("custom_ssl_certificates").Int()))
	}
	return &schema.CostComponent{
		Name:            "Dedicated IP custom SSLs",
		Unit:            "certificates",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Service:    strPtr("AmazonCloudFront"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", Value: strPtr("SSL-Cert-Custom")},
			},
		},
	}
}
