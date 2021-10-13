package aws

import (
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

type regionData struct {
	awsGroupedName string
	priceRegion    string
	usageKey       string
	showNil        bool
}

var (
	tierStarts = []int{0, 10240, 51200, 153600, 512000, 1048576, 5242880}
	tierLimits = []int{10240, 40960, 102400, 358400, 536576, 4194304}
	tierNames  = []string{"first 10TB", "next 40TB", "next 100TB", "next 350TB", "next 524TB", "next 4PB", "over 5PB"}

	regionsData = []*regionData{
		{
			awsGroupedName: "US, Mexico, Canada",
			priceRegion:    "United States",
			usageKey:       "us",
			showNil:        true,
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
)

func GetCloudfrontDistributionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudfront_distribution",
		RFunc: NewCloudfrontDistribution,
	}
}

func NewCloudfrontDistribution(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var components []*schema.CostComponent

	if v := encryptionRequests(d, u); v != nil {
		components = append(components, v)
	}

	if v := realtimeLogs(d, u); v != nil {
		components = append(components, v)
	}

	if v := customSSLCertificate(d, u); v != nil {
		components = append(components, v)
	}

	components = append(components, invalidationRequests(u)...)

	regionalUsage := newCloudfrontRegionUsage(u)
	subResources := regionalUsage.toSubResources()

	if v := shieldRequests(d, u); v != nil {
		subResources = append(subResources, v)
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: components,
		SubResources:   subResources,
	}
}

type cloudfrontRegionUsage struct {
	definitions map[string]struct{}

	MonthlyDataTransferToInternet map[string]gjson.Result
	MonthlyDataTransferToOrigin   map[string]gjson.Result
	MonthlyHTTPReq                map[string]gjson.Result
	MonthlyHTTPSReq               map[string]gjson.Result
}

func newCloudfrontRegionUsage(u *schema.UsageData) cloudfrontRegionUsage {
	regionUsage := cloudfrontRegionUsage{
		definitions:                   map[string]struct{}{},
		MonthlyDataTransferToInternet: map[string]gjson.Result{},
		MonthlyDataTransferToOrigin:   map[string]gjson.Result{},
		MonthlyHTTPReq:                map[string]gjson.Result{},
		MonthlyHTTPSReq:               map[string]gjson.Result{},
	}

	if u == nil {
		return regionUsage
	}

	regionUsage.MonthlyDataTransferToInternet = u.Get("monthly_data_transfer_to_internet_gb").Map()
	for k := range regionUsage.MonthlyDataTransferToInternet {
		regionUsage.definitions[k] = struct{}{}
	}

	regionUsage.MonthlyDataTransferToOrigin = u.Get("monthly_data_transfer_to_origin_gb").Map()
	for k := range regionUsage.MonthlyDataTransferToOrigin {
		regionUsage.definitions[k] = struct{}{}
	}

	regionUsage.MonthlyHTTPReq = u.Get("monthly_http_requests").Map()
	for k := range regionUsage.MonthlyHTTPReq {
		regionUsage.definitions[k] = struct{}{}
	}

	regionUsage.MonthlyHTTPSReq = u.Get("monthly_https_requests").Map()
	for k := range regionUsage.MonthlyHTTPSReq {
		regionUsage.definitions[k] = struct{}{}
	}

	return regionUsage
}

func (c cloudfrontRegionUsage) toSubResources() []*schema.Resource {
	var resources []*schema.Resource
	for _, regData := range regionsData {
		if _, ok := c.definitions[regData.usageKey]; !ok && !regData.showNil {
			continue
		}

		resource := &schema.Resource{
			Name: regData.awsGroupedName,
		}

		var components []*schema.CostComponent
		components = append(components, c.dataOutToInternet(regData)...)
		components = append(components, c.regionalDataOutToOrigin(regData))
		components = append(components, c.httpRequests(regData))
		components = append(components, c.httpsRequests(regData))

		resource.CostComponents = components
		resources = append(resources, resource)
	}

	return resources
}

func (c cloudfrontRegionUsage) dataOutToInternet(regData *regionData) []*schema.CostComponent {
	costName := "Data transfer out to internet"

	fromLocation := regData.priceRegion
	usageKey := regData.usageKey

	var quantity *decimal.Decimal
	if _, ok := c.MonthlyDataTransferToInternet[usageKey]; ok {
		quantity = decimalPtr(decimal.NewFromInt(c.MonthlyDataTransferToInternet[usageKey].Int()))
	}

	if quantity == nil {
		return []*schema.CostComponent{
			dataOutCostComponent(costName, tierNames[0], fromLocation, 0, nil),
		}
	}

	tiers := usage.CalculateTierBuckets(*quantity, tierLimits)
	var components []*schema.CostComponent
	for i := range tiers {
		if tiers[i].GreaterThan(decimal.Zero) {
			components = append(
				components,
				dataOutCostComponent(costName, tierNames[i], fromLocation, tierStarts[i], &tiers[i]),
			)
		}
	}

	return components
}

func dataOutCostComponent(costName, usageName, fromLocation string, startUsage int, quantity *decimal.Decimal) *schema.CostComponent {
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
	}
}

func (c cloudfrontRegionUsage) regionalDataOutToOrigin(regData *regionData) *schema.CostComponent {
	name := "Data transfer out to origin"

	apiRegion := regData.priceRegion
	usageKey := regData.usageKey
	var quantity *decimal.Decimal
	if _, ok := c.MonthlyDataTransferToOrigin[usageKey]; ok {
		quantity = decimalPtr(decimal.NewFromInt(c.MonthlyDataTransferToOrigin[usageKey].Int()))
	}

	return &schema.CostComponent{
		Name:            name,
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
	}
}

func (c cloudfrontRegionUsage) httpRequests(regData *regionData) *schema.CostComponent {
	name := "HTTP requests"

	apiRegion := regData.priceRegion
	usageKey := regData.usageKey
	var quantity *decimal.Decimal
	if _, ok := c.MonthlyHTTPReq[usageKey]; ok {
		quantity = decimalPtr(decimal.NewFromInt(c.MonthlyHTTPReq[usageKey].Int()))
	}

	return &schema.CostComponent{
		Name:            name,
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
	}
}

func (c cloudfrontRegionUsage) httpsRequests(regData *regionData) *schema.CostComponent {
	name := "HTTPS requests"

	apiRegion := regData.priceRegion
	usageKey := regData.usageKey
	var quantity *decimal.Decimal
	if _, ok := c.MonthlyHTTPSReq[usageKey]; ok {
		quantity = decimalPtr(decimal.NewFromInt(c.MonthlyHTTPSReq[usageKey].Int()))
	}

	return &schema.CostComponent{
		Name:            name,
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
	}
}

var regionShieldMapping = map[string]string{
	"us-gov-west-1":   "us",
	"us-gov-east-1":   "us",
	"us-east-1":       "us",
	"us-east-2":       "us",
	"us-west-1":       "us",
	"us-west-2":       "us",
	"us-west-2-lax-1": "us",
	"eu-central-1":    "europe",
	"eu-west-1":       "europe",
	"eu-west-2":       "europe",
	"eu-south-1":      "europe",
	"eu-west-3":       "europe",
	"eu-north-1":      "europe",
	"ap-northeast-1":  "japan",
	"ap-northeast-2":  "south_korea",
	"ap-southeast-1":  "singapore",
	"ap-southeast-2":  "australia",
	"ap-south-1":      "india",
	"sa-east-1":       "south_america",
}

func shieldRequests(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	if !d.Get("origin.0.origin_shield.0.enabled").Bool() {
		return nil
	}

	resource := &schema.Resource{
		Name: "Origin shield HTTP requests",
	}

	region := d.Get("region").String()
	if d.Get("origin.0.origin_shield.0.origin_shield_region").Exists() {
		region = d.Get("origin.0.origin_shield.0.origin_shield_region").String()
	}

	var apiRegion string
	if v, ok := regionMapping[region]; ok {
		apiRegion = v
	}

	if apiRegion == "" {
		log.Warnf("region %s not supported for origin shield requests skipping", region)
		return nil
	}

	// we need to find the legacy usage key that's defined in the usage file
	// this is done so we support backwards compatibility and don't nuke old usage files.
	var usageKey string
	if v, ok := regionShieldMapping[region]; ok {
		usageKey = v
	}

	if usageKey == "" {
		log.Warnf("region %s not supported for origin shield requests, unsupported in usage file, skipping", region)
		return nil
	}

	var quantity *decimal.Decimal
	if u != nil {
		shieldU := u.Get("monthly_shield_requests").Map()
		if _, ok := shieldU[usageKey]; ok {
			quantity = decimalPtr(decimal.NewFromInt(shieldU[usageKey].Int()))
		}
	}

	resource.CostComponents = []*schema.CostComponent{
		{
			Name:            apiRegion,
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
		},
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

func encryptionRequests(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	// for some reason field_level_encryption_id is set to a raw value of "null" when empty so
	// Exists() method returns true event thought the value is Type is Null
	if d.Get("default_cache_behavior.0.field_level_encryption_id").Type == gjson.Null {
		return nil
	}

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

func realtimeLogs(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	if !d.Get("logging_config.0.bucket").Exists() {
		return nil
	}

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

func customSSLCertificate(d *schema.ResourceData, u *schema.UsageData) *schema.CostComponent {
	if d.Get("viewer_certificate.0.ssl_support_method").String() != "vip" {
		return nil
	}

	quantity := decimalPtr(decimal.NewFromInt(1))
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
