package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type regionData struct {
	awsGroupedName string
	priceRegion    string
	usageKey       string
}

type usageFilterData struct {
	usageNumber int64
	usageName   string
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

	regionsData := []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "monthly_gb_transfer_internet.us",
		},
		{
			awsGroupedName: "Europe, Israel",
			priceRegion:    "Europe",
			usageKey:       "monthly_gb_transfer_internet.europe",
		},
		{
			awsGroupedName: "South Africa, Kenya, Middle East",
			priceRegion:    "South Africa",
			usageKey:       "monthly_gb_transfer_internet.south_africa",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "monthly_gb_transfer_internet.south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "monthly_gb_transfer_internet.japan",
		},
		{
			awsGroupedName: "Australia, New Zealand",
			priceRegion:    "Australia",
			usageKey:       "monthly_gb_transfer_internet.australia",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "monthly_gb_transfer_internet.asia_pacific",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "monthly_gb_transfer_internet.india",
		},
	}

	usageFilters := []*usageFilterData{
		{
			usageNumber: 10240,
			usageName:   "first 10TB",
		},
		{
			usageNumber: 51200,
			usageName:   "next 40TB",
		},
		{
			usageNumber: 153600,
			usageName:   "next 100TB",
		},
		{
			usageNumber: 512000,
			usageName:   "next 350TB",
		},
		{
			usageNumber: 1048576,
			usageName:   "next 524TB",
		},
		{
			usageNumber: 5242880,
			usageName:   "next 4PB",
		},
		{
			usageNumber: 0,
			usageName:   "over 5PB",
		},
	}

	// Because india has different usage amounts
	indiaUsageFilters := []*usageFilterData{
		{
			usageNumber: 10240,
			usageName:   "first 10TB",
		},
		{
			usageNumber: 51200,
			usageName:   "next 40TB",
		},
		{
			usageNumber: 153600,
			usageName:   "next 100TB",
		},
		{
			usageNumber: 0,
			usageName:   "over 150TB",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey

		var usage int64
		var used int64
		var lastEndUsageAmount int64
		if u != nil && u.Get(usageKey).Exists() {
			usage = u.Get(usageKey).Int()
		}

		selectedUsageFilters := usageFilters
		if apiRegion == "India" {
			selectedUsageFilters = indiaUsageFilters
		}
		for idx, usageFilter := range selectedUsageFilters {
			usageName := usageFilter.usageName
			endUsageAmount := usageFilter.usageNumber
			var quantity *decimal.Decimal
			if endUsageAmount != 0 && usage >= endUsageAmount {
				used = endUsageAmount - used
				lastEndUsageAmount = endUsageAmount
				quantity = decimalPtr(decimal.NewFromInt(used))
			} else if usage > lastEndUsageAmount {
				used = usage - lastEndUsageAmount
				lastEndUsageAmount = endUsageAmount
				quantity = decimalPtr(decimal.NewFromInt(used))
			}
			var usageFilter string
			if endUsageAmount != 0 {
				usageFilter = fmt.Sprint(endUsageAmount)
			} else {
				usageFilter = "Inf"
			}
			if quantity == nil && idx > 0 {
				continue
			}
			resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
				Name:            fmt.Sprintf("%v (%v)", awsRegion, usageName),
				Unit:            "GB",
				UnitMultiplier:  1,
				MonthlyQuantity: quantity,
				ProductFilter: &schema.ProductFilter{
					VendorName: strPtr("aws"),
					Service:    strPtr("AmazonCloudFront"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "transferType", Value: strPtr("CloudFront Outbound")},
						{Key: "fromLocation", Value: strPtr(apiRegion)},
					},
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr(usageFilter),
				},
			})
		}

	}

	return resource
}

func regionalDataOutToOrigin(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:           "Data transfer out to origin",
		CostComponents: []*schema.CostComponent{},
	}

	regionsData := []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "monthly_gb_transfer_origin.us",
		},
		{
			awsGroupedName: "Europe, Israel",
			priceRegion:    "Europe",
			usageKey:       "monthly_gb_transfer_origin.europe",
		},
		{
			awsGroupedName: "South Africa, Kenya, Middle East",
			priceRegion:    "South Africa",
			usageKey:       "monthly_gb_transfer_origin.south_africa",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "monthly_gb_transfer_origin.south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "monthly_gb_transfer_origin.japan",
		},
		{
			awsGroupedName: "Australia, New Zealand",
			priceRegion:    "Australia",
			usageKey:       "monthly_gb_transfer_origin.australia",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "monthly_gb_transfer_origin.asia_pacific",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "monthly_gb_transfer_origin.india",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey
		var quantity *decimal.Decimal
		if u != nil && u.Get(usageKey).Exists() {
			quantity = decimalPtr(decimal.NewFromInt(u.Get(usageKey).Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "GB",
			UnitMultiplier:  1,
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

	regionsData := []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "monthly_http_requests.us",
		},
		{
			awsGroupedName: "Europe, Israel",
			priceRegion:    "Europe",
			usageKey:       "monthly_http_requests.europe",
		},
		{
			awsGroupedName: "South Africa, Kenya, Middle East",
			priceRegion:    "South Africa",
			usageKey:       "monthly_http_requests.south_africa",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "monthly_http_requests.south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "monthly_http_requests.japan",
		},
		{
			awsGroupedName: "Australia, New Zealand",
			priceRegion:    "Australia",
			usageKey:       "monthly_http_requests.australia",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "monthly_http_requests.asia_pacific",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "monthly_http_requests.india",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey
		var quantity *decimal.Decimal
		if u != nil && u.Get(usageKey).Exists() {
			quantity = decimalPtr(decimal.NewFromInt(u.Get(usageKey).Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "requests",
			UnitMultiplier:  10000,
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

	regionsData := []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "monthly_https_requests.us",
		},
		{
			awsGroupedName: "Europe, Israel",
			priceRegion:    "Europe",
			usageKey:       "monthly_https_requests.europe",
		},
		{
			awsGroupedName: "South Africa, Kenya, Middle East",
			priceRegion:    "South Africa",
			usageKey:       "monthly_https_requests.south_africa",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "monthly_https_requests.south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "monthly_https_requests.japan",
		},
		{
			awsGroupedName: "Australia, New Zealand",
			priceRegion:    "Australia",
			usageKey:       "monthly_https_requests.australia",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "monthly_https_requests.asia_pacific",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "monthly_https_requests.india",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey
		var quantity *decimal.Decimal
		if u != nil && u.Get(usageKey).Exists() {
			quantity = decimalPtr(decimal.NewFromInt(u.Get(usageKey).Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "requests",
			UnitMultiplier:  10000,
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

	regionsData := []*regionData{
		{
			awsGroupedName: "US",
			priceRegion:    "US East (N. Virginia)",
			usageKey:       "monthly_shield_requests.us",
		},
		{
			awsGroupedName: "Europe",
			priceRegion:    "EU (Frankfurt)",
			usageKey:       "monthly_shield_requests.europe",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America (Sao Paulo)",
			usageKey:       "monthly_shield_requests.south_america",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Asia Pacific (Tokyo)",
			usageKey:       "monthly_shield_requests.japan",
		},
		{
			awsGroupedName: "Australia",
			priceRegion:    "Asia Pacific (Sydney)",
			usageKey:       "monthly_shield_requests.australia",
		},
		{
			awsGroupedName: "Singapore",
			priceRegion:    "Asia Pacific (Singapore)",
			usageKey:       "monthly_shield_requests.singapore",
		},
		{
			awsGroupedName: "South Korea",
			priceRegion:    "Asia Pacific (Seoul)",
			usageKey:       "monthly_shield_requests.south_korea",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "Asia Pacific (Mumbai)",
			usageKey:       "monthly_shield_requests.india",
		},
	}

	for _, regData := range regionsData {
		awsRegion := regData.awsGroupedName
		apiRegion := regData.priceRegion
		usageKey := regData.usageKey
		var quantity *decimal.Decimal
		if u != nil && u.Get(usageKey).Exists() {
			quantity = decimalPtr(decimal.NewFromInt(u.Get(usageKey).Int()))
		}
		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            awsRegion,
			Unit:            "requests",
			UnitMultiplier:  10000,
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
	if u != nil && u.Get("monthly_invalidation_paths").Exists() {
		usageAmount := u.Get("monthly_invalidation_paths").Int()
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
			UnitMultiplier:  1,
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
			UnitMultiplier:  1,
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
		Unit:            "requests",
		UnitMultiplier:  10000,
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
		Unit:            "lines",
		UnitMultiplier:  1000000,
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
	if u != nil && u.Get("monthly_custom_ssl_certificates").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("monthly_custom_ssl_certificates").Int()))
	}
	return &schema.CostComponent{
		Name:            "Dedicated IP custom SSLs",
		Unit:            "certificates",
		UnitMultiplier:  1,
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
