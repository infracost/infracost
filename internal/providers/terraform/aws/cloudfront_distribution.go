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
	return &schema.Resource{
		Name: d.Address,
		CostComponents: []*schema.CostComponent{
			encryptionRequests(u),
			realtimeLogs(u),
			customSSLCertificate(u),
		},
		SubResources: []*schema.Resource{
			invalidationPaths(u),
			regionalDataOutToInternet(u),
			regionalDataOutToOrigin(u),
			httpRequests(u),
			httpsRequests(u),
			shieldRequests(u),
		},
	}
}

func regionalDataOutToInternet(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:         "Regional data transfer out to internet",
		SubResources: []*schema.Resource{},
	}

	regionsData := []*regionData{
		{
			awsGroupedName: "United States, Mexico, & Canada",
			priceRegion:    "United States",
			usageKey:       "united_states_data_transfer_internet_gb",
		},
		{
			awsGroupedName: "Europe & Israel",
			priceRegion:    "Europe",
			usageKey:       "europe_data_transfer_internet_gb",
		},
		{
			awsGroupedName: "South Africa, Kenya, & Middle East",
			priceRegion:    "South Africa",
			usageKey:       "south_africa_data_transfer_internet_gb",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "south_america_data_transfer_internet_gb",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "japan_data_transfer_internet_gb",
		},
		{
			awsGroupedName: "Australia & New Zealand",
			priceRegion:    "Australia",
			usageKey:       "australia_data_transfer_internet_gb",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "asia_pacific_data_transfer_internet_gb",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "india_data_transfer_internet_gb",
		},
	}

	usageFilters := []*usageFilterData{
		{
			usageNumber: 10240,
			usageName:   "First 10TB",
		},
		{
			usageNumber: 51200,
			usageName:   "Next 40TB",
		},
		{
			usageNumber: 153600,
			usageName:   "Next 100TB",
		},
		{
			usageNumber: 512000,
			usageName:   "Next 350TB",
		},
		{
			usageNumber: 1048576,
			usageName:   "Next 524TB",
		},
		{
			usageNumber: 5242880,
			usageName:   "Next 4PB",
		},
		{
			usageNumber: 0,
			usageName:   "Over 5PB",
		},
	}

	// Because india has different usage amounts
	indiaUsageFilters := []*usageFilterData{
		{
			usageNumber: 10240,
			usageName:   "First 10TB",
		},
		{
			usageNumber: 51200,
			usageName:   "Next 40TB",
		},
		{
			usageNumber: 153600,
			usageName:   "Next 100TB",
		},
		{
			usageNumber: 0,
			usageName:   "Over 150TB",
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

		regionResource := &schema.Resource{
			Name:           awsRegion,
			CostComponents: []*schema.CostComponent{},
		}

		selectedUsageFilters := usageFilters
		if apiRegion == "India" {
			selectedUsageFilters = indiaUsageFilters
		}
		for _, usageFilter := range selectedUsageFilters {
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
			regionResource.CostComponents = append(regionResource.CostComponents, &schema.CostComponent{
				Name:            usageName,
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

		resource.SubResources = append(resource.SubResources, regionResource)

	}

	return resource
}

func regionalDataOutToOrigin(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:           "Regional data transfer out to origin",
		CostComponents: []*schema.CostComponent{},
	}

	regionsData := []*regionData{
		{
			awsGroupedName: "United States, Mexico, & Canada",
			priceRegion:    "United States",
			usageKey:       "united_states_data_transfer_origin_gb",
		},
		{
			awsGroupedName: "Europe & Israel",
			priceRegion:    "Europe",
			usageKey:       "europe_data_transfer_origin_gb",
		},
		{
			awsGroupedName: "South Africa, Kenya, & Middle East",
			priceRegion:    "South Africa",
			usageKey:       "south_africa_data_transfer_origin_gb",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "south_america_data_transfer_origin_gb",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "japan_data_transfer_origin_gb",
		},
		{
			awsGroupedName: "Australia & New Zealand",
			priceRegion:    "Australia",
			usageKey:       "australia_data_transfer_origin_gb",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "asia_pacific_data_transfer_origin_gb",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "india_data_transfer_origin_gb",
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
			awsGroupedName: "United States, Mexico, & Canada",
			priceRegion:    "United States",
			usageKey:       "united_states_http_requests",
		},
		{
			awsGroupedName: "Europe & Israel",
			priceRegion:    "Europe",
			usageKey:       "europe_http_requests",
		},
		{
			awsGroupedName: "South Africa, Kenya, & Middle East",
			priceRegion:    "South Africa",
			usageKey:       "south_africa_http_requests",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "south_america_http_requests",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "japan_http_requests",
		},
		{
			awsGroupedName: "Australia & New Zealand",
			priceRegion:    "Australia",
			usageKey:       "australia_http_requests",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "asia_pacific_http_requests",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "india_http_requests",
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
			awsGroupedName: "United States, Mexico, & Canada",
			priceRegion:    "United States",
			usageKey:       "united_states_https_requests",
		},
		{
			awsGroupedName: "Europe & Israel",
			priceRegion:    "Europe",
			usageKey:       "europe_https_requests",
		},
		{
			awsGroupedName: "South Africa, Kenya, & Middle East",
			priceRegion:    "South Africa",
			usageKey:       "south_africa_https_requests",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America",
			usageKey:       "south_america_https_requests",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Japan",
			usageKey:       "japan_https_requests",
		},
		{
			awsGroupedName: "Australia & New Zealand",
			priceRegion:    "Australia",
			usageKey:       "australia_https_requests",
		},
		{
			awsGroupedName: "Hong Kong, Philippines, Asia Pacific",
			priceRegion:    "Asia Pacific",
			usageKey:       "asia_pacific_https_requests",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "India",
			usageKey:       "india_https_requests",
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
			awsGroupedName: "United States",
			priceRegion:    "US East (N. Virginia)",
			usageKey:       "united_states_shield_requests",
		},
		{
			awsGroupedName: "Europe",
			priceRegion:    "EU (Frankfurt)",
			usageKey:       "europe_shield_requests",
		},
		{
			awsGroupedName: "South America",
			priceRegion:    "South America (Sao Paulo)",
			usageKey:       "south_america_shield_requests",
		},
		{
			awsGroupedName: "Japan",
			priceRegion:    "Asia Pacific (Tokyo)",
			usageKey:       "japan_shield_requests",
		},
		{
			awsGroupedName: "Australia",
			priceRegion:    "Asia Pacific (Sydney)",
			usageKey:       "australia_shield_requests",
		},
		{
			awsGroupedName: "Singapore",
			priceRegion:    "Asia Pacific (Singapore)",
			usageKey:       "singapore_shield_requests",
		},
		{
			awsGroupedName: "South Korea",
			priceRegion:    "Asia Pacific (Seoul)",
			usageKey:       "south_korea_shield_requests",
		},
		{
			awsGroupedName: "India",
			priceRegion:    "Asia Pacific (Mumbai)",
			usageKey:       "india_shield_requests",
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

func invalidationPaths(u *schema.UsageData) *schema.Resource {
	var freeQuantity *decimal.Decimal
	var paidQuantity *decimal.Decimal
	if u != nil && u.Get("invalidation_paths").Exists() {
		usageAmount := u.Get("invalidation_paths").Int()
		if usageAmount > 1000 {
			freeQuantity = decimalPtr(decimal.NewFromInt(1000))
			paidQuantity = decimalPtr(decimal.NewFromInt(usageAmount - 1000))
		} else {
			freeQuantity = decimalPtr(decimal.NewFromInt(usageAmount))
		}
	}

	return &schema.Resource{
		Name: "Invalidation requests",
		CostComponents: []*schema.CostComponent{
			{
				Name:            "First 1000 paths",
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
			{
				Name:            "Over 1000 paths",
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
			},
		},
	}
}

func encryptionRequests(u *schema.UsageData) *schema.CostComponent {
	var quantity *decimal.Decimal
	if u != nil && u.Get("encryption_requests").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("encryption_requests").Int()))
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
	if u != nil && u.Get("log_lines").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("log_lines").Int()))
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
	if u != nil && u.Get("custom_ssl_certificates").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("custom_ssl_certificates").Int()))
	}
	return &schema.CostComponent{
		Name:            "Dedicated IP custom SSL",
		Unit:            "certificate",
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
