package aws

import (
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/usage"
)

type CloudfrontDistribution struct {
	Address string
	Region  string

	IsOriginShieldEnabled     bool
	IsSSLSupportMethodVIP     bool
	HasLoggingConfigBucket    bool
	HasFieldLevelEncryptionID bool
	OriginShieldRegion        string

	// "usage" args
	MonthlyHTTPRequests             *cloudfrontDistributionRegionRequestsUsage     `infracost_usage:"monthly_http_requests"`
	MonthlyHTTPSRequests            *cloudfrontDistributionRegionRequestsUsage     `infracost_usage:"monthly_https_requests"`
	MonthlyShieldRequests           *cloudfrontDistributionShieldRequestsUsage     `infracost_usage:"monthly_shield_requests"`
	MonthlyInvalidationRequests     *int64                                         `infracost_usage:"monthly_invalidation_requests"`
	MonthlyEncryptionRequests       *int64                                         `infracost_usage:"monthly_encryption_requests"`
	MonthlyLogLines                 *int64                                         `infracost_usage:"monthly_log_lines"`
	MonthlyDataTransferToInternetGB *cloudfrontDistributionRegionDataTransferUsage `infracost_usage:"monthly_data_transfer_to_internet_gb"`
	MonthlyDataTransferToOriginGB   *cloudfrontDistributionRegionDataTransferUsage `infracost_usage:"monthly_data_transfer_to_origin_gb"`
	CustomSslCertificates           *int64                                         `infracost_usage:"custom_ssl_certificates"`
}

type cloudfrontDistributionRegionDataTransferUsage struct {
	US           *float64 `infracost_usage:"us"`
	Europe       *float64 `infracost_usage:"europe"`
	SouthAfrica  *float64 `infracost_usage:"south_africa"`
	SouthAmerica *float64 `infracost_usage:"south_america"`
	Japan        *float64 `infracost_usage:"japan"`
	Australia    *float64 `infracost_usage:"australia"`
	AsiaPacific  *float64 `infracost_usage:"asia_pacific"`
	India        *float64 `infracost_usage:"india"`
}

type cloudfrontDistributionRegionRequestsUsage struct {
	US           *int64 `infracost_usage:"us"`
	Europe       *int64 `infracost_usage:"europe"`
	SouthAfrica  *int64 `infracost_usage:"south_africa"`
	SouthAmerica *int64 `infracost_usage:"south_america"`
	Japan        *int64 `infracost_usage:"japan"`
	Australia    *int64 `infracost_usage:"australia"`
	AsiaPacific  *int64 `infracost_usage:"asia_pacific"`
	India        *int64 `infracost_usage:"india"`
}

type cloudfrontDistributionShieldRequestsUsage struct {
	US           *int64 `infracost_usage:"us"`
	Europe       *int64 `infracost_usage:"europe"`
	SouthAmerica *int64 `infracost_usage:"south_america"`
	Japan        *int64 `infracost_usage:"japan"`
	Australia    *int64 `infracost_usage:"australia"`
	Singapore    *int64 `infracost_usage:"singapore"`
	SouthKorea   *int64 `infracost_usage:"south_korea"`
	Indonesia    *int64 `infracost_usage:"indonesia"`
	India        *int64 `infracost_usage:"india"`
	MiddleEast   *int64 `infracost_usage:"middle_east"`
}

func (r *CloudfrontDistribution) CoreType() string {
	return "CloudfrontDistribution"
}

func (r *CloudfrontDistribution) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{
			Key:          "monthly_http_requests",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_http_requests", Items: cloudfrontDistributionRegionRequestsSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "monthly_https_requests",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_https_requests", Items: cloudfrontDistributionRegionRequestsSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "monthly_shield_requests",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_shield_requests", Items: cloudfrontDistributionShieldRequestsSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{Key: "monthly_invalidation_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_encryption_requests", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_log_lines", ValueType: schema.Int64, DefaultValue: 0},
		{
			Key:          "monthly_data_transfer_to_internet_gb",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_data_transfer_to_internet_gb", Items: cloudfrontDistributionRegionDataTransferSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{
			Key:          "monthly_data_transfer_to_origin_gb",
			DefaultValue: &usage.ResourceUsage{Name: "monthly_data_transfer_to_origin_gb", Items: cloudfrontDistributionRegionDataTransferSchema},
			ValueType:    schema.SubResourceUsage,
		},
		{Key: "custom_ssl_certificates", ValueType: schema.Int64, DefaultValue: 0},
	}
}

var cloudfrontDistributionRegionRequestsSchema = []*schema.UsageItem{
	{Key: "us", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "europe", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "south_africa", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "south_america", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "japan", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "australia", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "asia_pacific", DefaultValue: 0, ValueType: schema.Float64},
	{Key: "india", DefaultValue: 0, ValueType: schema.Float64},
}

var cloudfrontDistributionRegionDataTransferSchema = []*schema.UsageItem{
	{Key: "us", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "europe", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "south_africa", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "south_america", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "japan", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "australia", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "asia_pacific", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "india", DefaultValue: 0, ValueType: schema.Int64},
}

var cloudfrontDistributionShieldRequestsSchema = []*schema.UsageItem{
	{Key: "us", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "europe", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "south_america", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "japan", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "australia", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "singapore", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "south_korea", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "india", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "middle_east", DefaultValue: 0, ValueType: schema.Int64},
}

func (r *CloudfrontDistribution) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudfrontDistribution) BuildResource() *schema.Resource {
	var components []*schema.CostComponent

	if r.MonthlyHTTPRequests == nil {
		r.MonthlyHTTPRequests = &cloudfrontDistributionRegionRequestsUsage{}
	}
	if r.MonthlyHTTPSRequests == nil {
		r.MonthlyHTTPSRequests = &cloudfrontDistributionRegionRequestsUsage{}
	}
	if r.MonthlyShieldRequests == nil {
		r.MonthlyShieldRequests = &cloudfrontDistributionShieldRequestsUsage{}
	}
	if r.MonthlyDataTransferToInternetGB == nil {
		r.MonthlyDataTransferToInternetGB = &cloudfrontDistributionRegionDataTransferUsage{}
	}
	if r.MonthlyDataTransferToOriginGB == nil {
		r.MonthlyDataTransferToOriginGB = &cloudfrontDistributionRegionDataTransferUsage{}
	}

	components = append(components, r.encryptionRequestsCostComponents()...)
	components = append(components, r.realtimeLogsCostComponents()...)
	components = append(components, r.customSSLCertificateCostComponents()...)
	components = append(components, r.shieldRequestsCostComponents()...)
	components = append(components, r.invalidationRequestsCostComponents()...)

	subResources := r.buildSubresources()

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: components,
		SubResources:   subResources,
	}
}

type cloudfrontDistributionRegionData struct {
	awsGroupedName                  string
	priceRegion                     string
	monthlyHTTPRequests             *int64
	monthlyHTTPSRequests            *int64
	monthlyDataTransferToInternetGB *float64
	monthlyDataTransferToOriginGB   *float64
}

func (c *cloudfrontDistributionRegionData) HasUsage() bool {
	return c.monthlyHTTPRequests != nil || c.monthlyHTTPSRequests != nil ||
		c.monthlyDataTransferToInternetGB != nil || c.monthlyDataTransferToOriginGB != nil
}

func (r *CloudfrontDistribution) buildSubresources() []*schema.Resource {
	regionsData := []*cloudfrontDistributionRegionData{
		{
			awsGroupedName:                  "US, Mexico, Canada",
			priceRegion:                     "United States",
			monthlyHTTPRequests:             r.MonthlyHTTPRequests.US,
			monthlyHTTPSRequests:            r.MonthlyHTTPSRequests.US,
			monthlyDataTransferToInternetGB: r.MonthlyDataTransferToInternetGB.US,
			monthlyDataTransferToOriginGB:   r.MonthlyDataTransferToOriginGB.US,
		},
		{
			awsGroupedName:                  "Europe, Israel",
			priceRegion:                     "Europe",
			monthlyHTTPRequests:             r.MonthlyHTTPRequests.Europe,
			monthlyHTTPSRequests:            r.MonthlyHTTPSRequests.Europe,
			monthlyDataTransferToInternetGB: r.MonthlyDataTransferToInternetGB.Europe,
			monthlyDataTransferToOriginGB:   r.MonthlyDataTransferToOriginGB.Europe,
		},
		{
			awsGroupedName:                  "South Africa, Kenya, Middle East",
			priceRegion:                     "South Africa",
			monthlyHTTPRequests:             r.MonthlyHTTPRequests.SouthAfrica,
			monthlyHTTPSRequests:            r.MonthlyHTTPSRequests.SouthAfrica,
			monthlyDataTransferToInternetGB: r.MonthlyDataTransferToInternetGB.SouthAfrica,
			monthlyDataTransferToOriginGB:   r.MonthlyDataTransferToOriginGB.SouthAfrica,
		},
		{
			awsGroupedName:                  "South America",
			priceRegion:                     "South America",
			monthlyHTTPRequests:             r.MonthlyHTTPRequests.SouthAmerica,
			monthlyHTTPSRequests:            r.MonthlyHTTPSRequests.SouthAmerica,
			monthlyDataTransferToInternetGB: r.MonthlyDataTransferToInternetGB.SouthAmerica,
			monthlyDataTransferToOriginGB:   r.MonthlyDataTransferToOriginGB.SouthAmerica,
		},
		{
			awsGroupedName:                  "Japan",
			priceRegion:                     "Japan",
			monthlyHTTPRequests:             r.MonthlyHTTPRequests.Japan,
			monthlyHTTPSRequests:            r.MonthlyHTTPSRequests.Japan,
			monthlyDataTransferToInternetGB: r.MonthlyDataTransferToInternetGB.Japan,
			monthlyDataTransferToOriginGB:   r.MonthlyDataTransferToOriginGB.Japan,
		},
		{
			awsGroupedName:                  "Australia, New Zealand",
			priceRegion:                     "Australia",
			monthlyHTTPRequests:             r.MonthlyHTTPRequests.Australia,
			monthlyHTTPSRequests:            r.MonthlyHTTPSRequests.Australia,
			monthlyDataTransferToInternetGB: r.MonthlyDataTransferToInternetGB.Australia,
			monthlyDataTransferToOriginGB:   r.MonthlyDataTransferToOriginGB.Australia,
		},
		{
			awsGroupedName:                  "Hong Kong, Philippines, Asia Pacific",
			priceRegion:                     "Asia Pacific",
			monthlyHTTPRequests:             r.MonthlyHTTPRequests.AsiaPacific,
			monthlyHTTPSRequests:            r.MonthlyHTTPSRequests.AsiaPacific,
			monthlyDataTransferToInternetGB: r.MonthlyDataTransferToInternetGB.AsiaPacific,
			monthlyDataTransferToOriginGB:   r.MonthlyDataTransferToOriginGB.AsiaPacific,
		},
		{
			awsGroupedName:                  "India",
			priceRegion:                     "India",
			monthlyHTTPRequests:             r.MonthlyHTTPRequests.India,
			monthlyHTTPSRequests:            r.MonthlyHTTPSRequests.India,
			monthlyDataTransferToInternetGB: r.MonthlyDataTransferToInternetGB.India,
			monthlyDataTransferToOriginGB:   r.MonthlyDataTransferToOriginGB.India,
		},
	}

	subresources := []*schema.Resource{}

	for _, data := range regionsData {
		if !data.HasUsage() {
			continue
		}

		subresources = append(subresources, r.buildRegionSubresource(data))
	}

	if len(subresources) == 0 {
		subresources = append(subresources, r.buildRegionSubresource(regionsData[0]))
	}

	return subresources
}

func (r *CloudfrontDistribution) buildRegionSubresource(regionData *cloudfrontDistributionRegionData) *schema.Resource {
	resource := &schema.Resource{
		Name: regionData.awsGroupedName,
	}

	components := []*schema.CostComponent{}
	components = append(components, r.dataOutToInternetCostComponents(regionData)...)
	components = append(components, r.dataOutToOriginCostComponents(regionData)...)
	components = append(components, r.httpRequestsCostComponents(regionData)...)
	components = append(components, r.httpsRequestsCostComponents(regionData)...)

	resource.CostComponents = components

	return resource
}

func (r *CloudfrontDistribution) dataOutToInternetCostComponents(regionData *cloudfrontDistributionRegionData) []*schema.CostComponent {
	tierStarts := []int{0, 10240, 51200, 153600, 512000, 1048576, 5242880}
	tierLimits := []int{10240, 40960, 102400, 358400, 536576, 4194304}
	tierNames := []string{"first 10TB", "next 40TB", "next 100TB", "next 350TB", "next 524TB", "next 4PB", "over 5PB"}

	fromLocation := regionData.priceRegion

	var quantity *decimal.Decimal
	if regionData.monthlyDataTransferToInternetGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*regionData.monthlyDataTransferToInternetGB))
	}

	if quantity == nil {
		return []*schema.CostComponent{
			r.buildDataOutCostComponent(tierNames[0], fromLocation, 0, nil),
		}
	}

	tiers := usage.CalculateTierBuckets(*quantity, tierLimits)
	var components []*schema.CostComponent
	for i := range tiers {
		if tiers[i].GreaterThan(decimal.Zero) {
			components = append(
				components,
				r.buildDataOutCostComponent(tierNames[i], fromLocation, tierStarts[i], &tiers[i]),
			)
		}
	}

	return components
}

func (r *CloudfrontDistribution) buildDataOutCostComponent(usageName, fromLocation string, startUsage int, quantity *decimal.Decimal) *schema.CostComponent {
	costName := "Data transfer out to internet"

	return &schema.CostComponent{
		Name:            fmt.Sprintf("%s (%s)", costName, usageName),
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
		UsageBased: true,
	}
}

func (r *CloudfrontDistribution) dataOutToOriginCostComponents(regionData *cloudfrontDistributionRegionData) []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	apiRegion := regionData.priceRegion

	var quantity *decimal.Decimal

	if regionData.monthlyDataTransferToOriginGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*regionData.monthlyDataTransferToOriginGB))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Data transfer out to origin",
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
		UsageBased: true,
	})

	return costComponents
}

func (r *CloudfrontDistribution) httpRequestsCostComponents(regionData *cloudfrontDistributionRegionData) []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	apiRegion := regionData.priceRegion

	var quantity *decimal.Decimal
	if regionData.monthlyHTTPRequests != nil {
		quantity = decimalPtr(decimal.NewFromInt(*regionData.monthlyHTTPRequests))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "HTTP requests",
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
		UsageBased: true,
	})

	return costComponents
}

func (r *CloudfrontDistribution) httpsRequestsCostComponents(regionData *cloudfrontDistributionRegionData) []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	apiRegion := regionData.priceRegion

	var quantity *decimal.Decimal
	if regionData.monthlyHTTPSRequests != nil {
		quantity = decimalPtr(decimal.NewFromInt(*regionData.monthlyHTTPSRequests))
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "HTTPS requests",
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
		UsageBased: true,
	})

	return costComponents
}

// See https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/origin-shield.html#choose-origin-shield-region for the list of regions that are supported for Origin Shield.
var regionShieldMapping = map[string]string{
	"us-gov-west-1":   "us",
	"us-gov-east-1":   "us",
	"us-east-1":       "us",
	"us-east-2":       "us",
	"us-west-1":       "us",
	"us-west-2":       "us",
	"us-west-2-lax-1": "us",
	"ca-central-1":    "us",
	"ap-south-1":      "india",
	"me-south-1":      "india",
	"ap-northeast-1":  "japan",
	"ap-northeast-2":  "south_korea",
	"ap-southeast-1":  "singapore",
	"ap-east-1":       "singapore",
	"ap-southeast-2":  "australia",
	"eu-central-1":    "europe",
	"eu-west-1":       "europe",
	"eu-west-2":       "europe",
	"eu-south-1":      "europe",
	"eu-west-3":       "europe",
	"eu-north-1":      "europe",
	"af-south-1":      "europe",
	"sa-east-1":       "south_america",
	"me-central-1":    "middle_east",
}

func (r *CloudfrontDistribution) shieldRequestsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.IsOriginShieldEnabled {
		return costComponents
	}

	region := r.Region
	if r.OriginShieldRegion != "" {
		region = r.OriginShieldRegion
	}

	var apiRegion string
	if v, ok := RegionMapping[region]; ok {
		apiRegion = v
	}

	if apiRegion == "" {
		logging.Logger.Warn().Msgf("Skipping Origin shield HTTP requests for resource %s. Could not find mapping for region %s", r.Address, region)
		return costComponents
	}

	var usageKey string
	if v, ok := regionShieldMapping[region]; ok {
		usageKey = v
	}

	if usageKey == "" {
		logging.Logger.Warn().Msgf("No usage for Origin shield HTTP requests for resource %s.  Region %s not supported in usage file.", r.Address, region)
	}

	regionData := map[string]*int64{
		"us":            r.MonthlyShieldRequests.US,
		"europe":        r.MonthlyShieldRequests.Europe,
		"south_america": r.MonthlyShieldRequests.SouthAmerica,
		"japan":         r.MonthlyShieldRequests.Japan,
		"australia":     r.MonthlyShieldRequests.Australia,
		"singapore":     r.MonthlyShieldRequests.Singapore,
		"south_korea":   r.MonthlyShieldRequests.SouthKorea,
		"indonesia":     r.MonthlyShieldRequests.Indonesia,
		"india":         r.MonthlyShieldRequests.India,
		"middle_east":   r.MonthlyShieldRequests.MiddleEast,
	}

	var quantity *decimal.Decimal
	if _, ok := regionData[usageKey]; ok && regionData[usageKey] != nil {
		quantity = decimalPtr(decimal.NewFromInt(*regionData[usageKey]))
	}

	pieces := strings.Split(apiRegion, "(")
	prettyName := strings.TrimSpace(pieces[0]) + " " + region

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            fmt.Sprintf("Origin shield HTTP requests (%s)", prettyName),
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
		UsageBased: true,
	})

	return costComponents
}

func (r *CloudfrontDistribution) invalidationRequestsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	var freeQuantity *decimal.Decimal
	var paidQuantity *decimal.Decimal
	if r.MonthlyInvalidationRequests != nil {
		usageAmount := *r.MonthlyInvalidationRequests
		if usageAmount < 1000 {
			freeQuantity = decimalPtr(decimal.NewFromInt(usageAmount))
		} else {
			freeQuantity = decimalPtr(decimal.NewFromInt(1000))
			paidQuantity = decimalPtr(decimal.NewFromInt(usageAmount - 1000))
		}
	}

	costComponents = append(costComponents, &schema.CostComponent{
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
		UsageBased: true,
	})

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
			UsageBased: true,
		})
	}

	return costComponents
}

func (r *CloudfrontDistribution) encryptionRequestsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.HasFieldLevelEncryptionID {
		return costComponents
	}

	var quantity *decimal.Decimal
	if r.MonthlyEncryptionRequests != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyEncryptionRequests))
	}

	costComponents = append(costComponents, &schema.CostComponent{
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
		UsageBased: true,
	})

	return costComponents
}

func (r *CloudfrontDistribution) realtimeLogsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.HasLoggingConfigBucket {
		return costComponents
	}

	var quantity *decimal.Decimal
	if r.MonthlyLogLines != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.MonthlyLogLines))
	}

	costComponents = append(costComponents, &schema.CostComponent{
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
		UsageBased: true,
	})

	return costComponents
}

func (r *CloudfrontDistribution) customSSLCertificateCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.IsSSLSupportMethodVIP {
		return costComponents
	}

	quantity := decimalPtr(decimal.NewFromInt(1))
	if r.CustomSslCertificates != nil {
		quantity = decimalPtr(decimal.NewFromInt(*r.CustomSslCertificates))
	}

	costComponents = append(costComponents, &schema.CostComponent{
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
		UsageBased: true,
	})

	return costComponents
}
