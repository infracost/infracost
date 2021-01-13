package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

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
			invalidationURLs(u),
			encryptionRequests(u),
			realtimeLogs(u),
		},
		SubResources: []*schema.Resource{
			regionalDataOutToOrigin(u),
			requests(u),
			shieldRequests(u),
		},
	}
}

func regionalDataOutToOrigin(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:           "Regional data transfer out to origin",
		CostComponents: []*schema.CostComponent{},
	}

	// regionMap structure is: aws grouped name -> [pricing region , usage data key]
	regionsMap := map[string][]string{
		"United States, Mexico, & Canada":      {"United States", "united_states_data_transfer_origin_gb"},
		"Europe & Israel":                      {"Europe", "europe_data_transfer_origin_gb"},
		"South Africa, Kenya, & Middle East":   {"South Africa", "south_africa_data_transfer_origin_gb"},
		"South America":                        {"South America", "south_america_data_transfer_origin_gb"},
		"Japan":                                {"Japan", "japan_data_transfer_origin_gb"},
		"Australia & New Zealand":              {"Australia", "australia_data_transfer_origin_gb"},
		"Hong Kong, Philippines, Asia Pacific": {"Asia Pacific", "asia_pacific_data_transfer_origin_gb"},
		"India":                                {"India", "india_data_transfer_origin_gb"},
	}

	for key, value := range regionsMap {
		awsRegion := key
		apiRegion := value[0]
		usageKey := value[1]
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

func requests(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name: "Request pricing for all http methods",
		SubResources: []*schema.Resource{
			httpRequests(u),
			httpsRequests(u),
		},
	}

	return resource
}

func httpRequests(u *schema.UsageData) *schema.Resource {
	resource := &schema.Resource{
		Name:         "HTTP requests",
		SubResources: []*schema.Resource{},
	}

	// regionMap structure is: aws grouped name -> [pricing region , usage data key]
	regionsMap := map[string][]string{
		"United States, Mexico, & Canada":      {"United States", "united_states_http_requests"},
		"Europe & Israel":                      {"Europe", "europe_http_requests"},
		"South Africa, Kenya, & Middle East":   {"South Africa", "south_africa_http_requests"},
		"South America":                        {"South America", "south_america_http_requests"},
		"Japan":                                {"Japan", "japan_http_requests"},
		"Australia & New Zealand":              {"Australia", "australia_http_requests"},
		"Hong Kong, Philippines, Asia Pacific": {"Asia Pacific", "asia_pacific_http_requests"},
		"India":                                {"India", "india_http_requests"},
	}

	for key, value := range regionsMap {
		awsRegion := key
		apiRegion := value[0]
		usageKey := value[1]
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

	// regionMap structure is: aws grouped name -> [pricing region , usage data key]
	regionsMap := map[string][]string{
		"United States, Mexico, & Canada":      {"United States", "united_states_https_requests"},
		"Europe & Israel":                      {"Europe", "europe_https_requests"},
		"South Africa, Kenya, & Middle East":   {"South Africa", "south_africa_https_requests"},
		"South America":                        {"South America", "south_america_https_requests"},
		"Japan":                                {"Japan", "japan_https_requests"},
		"Australia & New Zealand":              {"Australia", "australia_https_requests"},
		"Hong Kong, Philippines, Asia Pacific": {"Asia Pacific", "asia_pacific_https_requests"},
		"India":                                {"India", "india_https_requests"},
	}

	for key, value := range regionsMap {
		awsRegion := key
		apiRegion := value[0]
		usageKey := value[1]
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
		Name:           "Origin shield request pricing for all http methods",
		CostComponents: []*schema.CostComponent{},
	}

	// regionMap structure is: aws grouped name -> [pricing region , usage data key]
	regionsMap := map[string][]string{
		"United States": {"US East (N. Virginia)", "united_states_shield_requests"},
		"Europe":        {"EU (Frankfurt)", "europe_shield_requests"},
		"South America": {"South America (Sao Paulo)", "south_america_shield_requests"},
		"Japan":         {"Asia Pacific (Tokyo)", "japan_shield_requests"},
		"Australia":     {"Asia Pacific (Sydney)", "australia_shield_requests"},
		"Singapore":     {"Asia Pacific (Singapore)", "singapore_shield_requests"},
		"South Korea":   {"Asia Pacific (Seoul)", "south_korea_shield_requests"},
		"India":         {"Asia Pacific (Mumbai)", "india_shield_requests"},
	}

	for key, value := range regionsMap {
		awsRegion := key
		apiRegion := value[0]
		usageKey := value[1]
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

func invalidationURLs(u *schema.UsageData) *schema.CostComponent {
	var quantity *decimal.Decimal
	if u != nil && u.Get("invalidation_requests").Exists() {
		quantity = decimalPtr(decimal.NewFromInt(u.Get("invalidation_requests").Int()))
	}
	return &schema.CostComponent{
		Name:            "Invalidation requests",
		Unit:            "urls",
		UnitMultiplier:  1,
		MonthlyQuantity: quantity,
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
