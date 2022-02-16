package google

import (
	"fmt"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

type NetworkEgressUsage struct {
	Address    string
	Region     string
	PrefixName string
}

type StorageBucketNetworkEgressUsage struct {
	SameContinent *float64 `infracost_usage:"same_continent"`
	Asia          *float64 `infracost_usage:"asia"`
	Worldwide     *float64 `infracost_usage:"worldwide"`
	China         *float64 `infracost_usage:"china"`
	Australia     *float64 `infracost_usage:"australia"`

	NetworkEgressUsage
}

var StorageBucketNetworkEgressUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Float64, DefaultValue: 0, Key: "same_continent"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "worldwide"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "china"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "australia"},
}

type ContainerRegistryNetworkEgressUsage struct {
	SameContinent *float64 `infracost_usage:"same_continent"`
	Asia          *float64 `infracost_usage:"asia"`
	Worldwide     *float64 `infracost_usage:"worldwide"`
	China         *float64 `infracost_usage:"china"`
	Australia     *float64 `infracost_usage:"australia"`

	NetworkEgressUsage
}

var ContainerRegistryNetworkEgressUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Float64, DefaultValue: 0, Key: "same_continent"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "worldwide"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "china"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "australia"},
}

type ComputeVPNGatewayNetworkEgressUsage struct {
	SameRegion   *float64 `infracost_usage:"same_region"`
	USOrCanada   *float64 `infracost_usage:"us_or_canada"`
	Europe       *float64 `infracost_usage:"europe"`
	Asia         *float64 `infracost_usage:"asia"`
	SouthAmerica *float64 `infracost_usage:"south_america"`
	Oceania      *float64 `infracost_usage:"oceania"`
	Worldwide    *float64 `infracost_usage:"worldwide"`

	NetworkEgressUsage
}

var ComputeVPNGatewayNetworkEgressUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Float64, DefaultValue: 0, Key: "same_region"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "us_or_canada"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "europe"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "south_america"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "oceania"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "worldwide"},
}

type ComputeExternalVPNGatewayNetworkEgressUsage struct {
	Asia      *float64 `infracost_usage:"asia"`
	Worldwide *float64 `infracost_usage:"worldwide"`
	China     *float64 `infracost_usage:"china"`
	Australia *float64 `infracost_usage:"australia"`

	NetworkEgressUsage
}

var ComputeExternalVPNGatewayNetworkEgressUsageSchema = []*schema.UsageItem{
	{ValueType: schema.Float64, DefaultValue: 0, Key: "asia"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "worldwide"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "china"},
	{ValueType: schema.Float64, DefaultValue: 0, Key: "australia"},
}

type egressRegionData struct {
	gRegion             string
	apiDescription      string
	apiDescriptionRegex string
	usageKey            string
	fixedRegion         string
}

type egressRegionUsageFilterData struct {
	usageNumber float64
	usageName   string
}

func (r *StorageBucketNetworkEgressUsage) BuildResource() *schema.Resource {
	regionsData := r.getEgressRegionsData()
	usageFiltersData := r.getEgressUsageFiltersData()
	defaultAPIRegionName := r.getEgressAPIRegionName()
	serviceName := r.getEgressAPIServiceName()

	resource := &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{},
	}

	// Same continent
	var quantity *decimal.Decimal
	if r.SameContinent != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.SameContinent))
	}
	resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
		Name:            fmt.Sprintf("%s in same continent", r.PrefixName),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr("global"),
			Service:    strPtr("Cloud Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("Inter-region GCP Storage egress within EU")},
			},
		},
	})

	for _, regData := range regionsData {
		usageKey := regData.usageKey
		usage := GetFloatFieldValueByUsageTag(usageKey, *r)
		newCostComponents := stepPricingHelper(usage, usageFiltersData, regData, defaultAPIRegionName, serviceName)
		resource.CostComponents = append(resource.CostComponents, newCostComponents...)
	}

	return resource
}

func (r *ContainerRegistryNetworkEgressUsage) BuildResource() *schema.Resource {
	regionsData := r.getEgressRegionsData()
	usageFiltersData := r.getEgressUsageFiltersData()
	defaultAPIRegionName := r.getEgressAPIRegionName()
	serviceName := r.getEgressAPIServiceName()

	resource := &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{},
	}

	// Same continent
	var quantity *decimal.Decimal
	if r.SameContinent != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.SameContinent))
	}
	resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
		Name:            fmt.Sprintf("%s in same continent", r.PrefixName),
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("gcp"),
			Region:     strPtr("global"),
			Service:    strPtr("Cloud Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("Inter-region GCP Storage egress within EU")},
			},
		},
	})

	for _, regData := range regionsData {
		usageKey := regData.usageKey
		usage := GetFloatFieldValueByUsageTag(usageKey, *r)
		newCostComponents := stepPricingHelper(usage, usageFiltersData, regData, defaultAPIRegionName, serviceName)
		resource.CostComponents = append(resource.CostComponents, newCostComponents...)
	}

	return resource
}

func (r *ComputeVPNGatewayNetworkEgressUsage) BuildResource() *schema.Resource {
	regionsData := r.getEgressRegionsData()
	usageFiltersData := r.getEgressUsageFiltersData()
	defaultAPIRegionName := r.getEgressAPIRegionName()
	serviceName := r.getEgressAPIServiceName()

	resource := &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{},
	}

	for _, regData := range regionsData {
		usageKey := regData.usageKey
		usage := GetFloatFieldValueByUsageTag(usageKey, *r)
		newCostComponents := stepPricingHelper(usage, usageFiltersData, regData, defaultAPIRegionName, serviceName)
		resource.CostComponents = append(resource.CostComponents, newCostComponents...)
	}

	return resource
}

func (r *ComputeExternalVPNGatewayNetworkEgressUsage) BuildResource() *schema.Resource {
	regionsData := r.getEgressRegionsData()
	usageFiltersData := r.getEgressUsageFiltersData()
	defaultAPIRegionName := r.getEgressAPIRegionName()
	serviceName := r.getEgressAPIServiceName()

	resource := &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{},
	}

	for _, regData := range regionsData {
		usageKey := regData.usageKey
		usage := GetFloatFieldValueByUsageTag(usageKey, *r)
		newCostComponents := stepPricingHelper(usage, usageFiltersData, regData, defaultAPIRegionName, serviceName)
		resource.CostComponents = append(resource.CostComponents, newCostComponents...)
	}

	return resource
}

func stepPricingHelper(usage float64, usageFiltersData []*egressRegionUsageFilterData, regData *egressRegionData, defaultAPIRegionName, serviceName *string) []*schema.CostComponent {
	costComponents := make([]*schema.CostComponent, 0)
	// TODO: Reformat to use tier helpers.
	var used float64
	var lastEndUsageAmount float64

	for idx, usageFilter := range usageFiltersData {
		usageName := usageFilter.usageName
		endUsageAmount := usageFilter.usageNumber
		var quantity *decimal.Decimal
		if endUsageAmount != 0 && usage >= endUsageAmount {
			used = endUsageAmount - used
			lastEndUsageAmount = endUsageAmount
			quantity = decimalPtr(decimal.NewFromFloat(used))
		} else if usage > lastEndUsageAmount {
			used = usage - lastEndUsageAmount
			lastEndUsageAmount = endUsageAmount
			quantity = decimalPtr(decimal.NewFromFloat(used))
		}
		var usageFilter string
		if endUsageAmount != 0 {
			usageFilter = fmt.Sprint(endUsageAmount)
		} else {
			usageFilter = ""
		}
		if quantity == nil && idx > 0 {
			continue
		}
		var apiRegion *string
		if regData.fixedRegion != "" {
			apiRegion = strPtr(regData.fixedRegion)
		} else {
			apiRegion = defaultAPIRegionName
		}
		var name string
		if usageName != "" {
			name = fmt.Sprintf("%v (%v)", regData.gRegion, usageName)
		} else {
			name = fmt.Sprintf("%v", regData.gRegion)
		}
		attributeFilters := make([]*schema.AttributeFilter, 0)
		if regData.apiDescriptionRegex != "" {
			attributeFilters = append(attributeFilters, &schema.AttributeFilter{Key: "description", ValueRegex: strPtr(regData.apiDescriptionRegex)})
		} else {
			attributeFilters = append(attributeFilters, &schema.AttributeFilter{Key: "description", Value: strPtr(regData.apiDescription)})
		}
		costComponents = append(costComponents, &schema.CostComponent{
			Name:            name,
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: quantity,
			ProductFilter: &schema.ProductFilter{
				Region:           apiRegion,
				VendorName:       strPtr("gcp"),
				Service:          serviceName,
				AttributeFilters: attributeFilters,
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr(usageFilter),
			},
		})
	}
	return costComponents
}

// ****** getEgressRegionsData funcs ******

func (r *StorageBucketNetworkEgressUsage) getEgressRegionsData() []*egressRegionData {
	return []*egressRegionData{
		{
			gRegion:        fmt.Sprintf("%s to worldwide excluding Asia, Australia", r.PrefixName),
			apiDescription: "Download Worldwide Destinations (excluding Asia & Australia)",
			usageKey:       "worldwide",
		},
		{
			gRegion:        fmt.Sprintf("%s to Asia excluding China, but including Hong Kong", r.PrefixName),
			apiDescription: "Download APAC",
			usageKey:       "asia",
		},
		{
			gRegion:        fmt.Sprintf("%s to China excluding Hong Kong", r.PrefixName),
			apiDescription: "Download China",
			usageKey:       "china",
		},
		{
			gRegion:        fmt.Sprintf("%s to Australia", r.PrefixName),
			apiDescription: "Download Australia",
			usageKey:       "australia",
		},
	}
}

func (r *ContainerRegistryNetworkEgressUsage) getEgressRegionsData() []*egressRegionData {
	return []*egressRegionData{
		{
			gRegion:        fmt.Sprintf("%s to worldwide excluding Asia, Australia", r.PrefixName),
			apiDescription: "Download Worldwide Destinations (excluding Asia & Australia)",
			usageKey:       "worldwide",
		},
		{
			gRegion:        fmt.Sprintf("%s to Asia excluding China, but including Hong Kong", r.PrefixName),
			apiDescription: "Download APAC",
			usageKey:       "asia",
		},
		{
			gRegion:        fmt.Sprintf("%s to China excluding Hong Kong", r.PrefixName),
			apiDescription: "Download China",
			usageKey:       "china",
		},
		{
			gRegion:        fmt.Sprintf("%s to Australia", r.PrefixName),
			apiDescription: "Download Australia",
			usageKey:       "australia",
		},
	}
}

func (r *ComputeVPNGatewayNetworkEgressUsage) getEgressRegionsData() []*egressRegionData {
	return []*egressRegionData{
		{
			gRegion: fmt.Sprintf("%s within the same region", r.PrefixName),
			// There is no same region option in APIs, so we always take this price in us-central1 region.
			apiDescription: "Network Vpn Inter Region Egress from Americas to Americas",
			usageKey:       "same_region",
			fixedRegion:    "us-central1",
		},
		{
			gRegion:        fmt.Sprintf("%s within the US or Canada", r.PrefixName),
			apiDescription: "Network Vpn Inter Region Egress from Americas to Montreal",
			usageKey:       "us_or_canada",
			fixedRegion:    "us-central1",
		},
		{
			gRegion:        fmt.Sprintf("%s within Europe", r.PrefixName),
			apiDescription: "Network Vpn Inter Region Egress from EMEA to EMEA",
			usageKey:       "europe",
			fixedRegion:    "europe-west1",
		},
		{
			gRegion:        fmt.Sprintf("%s within Asia", r.PrefixName),
			apiDescription: "Network Vpn Inter Region Egress from Japan to Seoul",
			usageKey:       "asia",
			fixedRegion:    "asia-northeast1",
		},
		{
			gRegion:        fmt.Sprintf("%s within South America", r.PrefixName),
			apiDescription: "Network Vpn Inter Region Egress from Sao Paulo to Sao Paulo",
			usageKey:       "south_america",
			fixedRegion:    "southamerica-east1",
		},
		{
			gRegion:        fmt.Sprintf("%s to/from Indonesia and Oceania", r.PrefixName),
			apiDescription: "Network Vpn Inter Region Egress from Sydney to Jakarta",
			usageKey:       "oceania",
			fixedRegion:    "australia-southeast1",
		},
		{
			gRegion:        fmt.Sprintf("%s between continents (excludes Oceania)", r.PrefixName),
			apiDescription: "Network Vpn Inter Region Egress from Finland to Singapore",
			usageKey:       "worldwide",
			fixedRegion:    "europe-north1",
		},
	}
}

func (r *ComputeExternalVPNGatewayNetworkEgressUsage) getEgressRegionsData() []*egressRegionData {
	return []*egressRegionData{
		{
			gRegion: fmt.Sprintf("%s to worldwide excluding China, Australia but including Hong Kong", r.PrefixName),
			// There is no worldwide option in APIs, so we take a random region.
			apiDescriptionRegex: "/Vpn Internet Egress .* to Americas/",
			usageKey:            "worldwide",
		},
		{
			gRegion:             fmt.Sprintf("%s to China excluding Hong Kong", r.PrefixName),
			apiDescriptionRegex: "/Vpn Internet Egress .* to China/",
			usageKey:            "china",
		},
		{
			gRegion:             fmt.Sprintf("%s to Australia", r.PrefixName),
			apiDescriptionRegex: "/Vpn Internet Egress .* to Australia/",
			usageKey:            "australia",
		},
	}
}

// ****** getEgressUsageFiltersData funcs ******

func (r *StorageBucketNetworkEgressUsage) getEgressUsageFiltersData() []*egressRegionUsageFilterData {
	return []*egressRegionUsageFilterData{
		{
			usageName:   "first 1TB",
			usageNumber: 1024,
		},
		{
			usageName:   "next 9TB",
			usageNumber: 10240,
		},
		{
			usageName:   "over 10TB",
			usageNumber: 0,
		},
	}
}

func (r *ContainerRegistryNetworkEgressUsage) getEgressUsageFiltersData() []*egressRegionUsageFilterData {
	return []*egressRegionUsageFilterData{
		{
			usageName:   "first 1TB",
			usageNumber: 1024,
		},
		{
			usageName:   "next 9TB",
			usageNumber: 10240,
		},
		{
			usageName:   "over 10TB",
			usageNumber: 0,
		},
	}
}

func (r *ComputeVPNGatewayNetworkEgressUsage) getEgressUsageFiltersData() []*egressRegionUsageFilterData {
	return []*egressRegionUsageFilterData{
		{
			usageNumber: 0,
		},
	}
}

func (r *ComputeExternalVPNGatewayNetworkEgressUsage) getEgressUsageFiltersData() []*egressRegionUsageFilterData {
	return []*egressRegionUsageFilterData{
		{
			usageName:   "first 1TB",
			usageNumber: 1024,
		},
		{
			usageName:   "next 9TB",
			usageNumber: 10240,
		},
		{
			usageName:   "over 10TB",
			usageNumber: 0,
		},
	}
}

// ****** getEgressAPIRegionName funcs ******

func (r *StorageBucketNetworkEgressUsage) getEgressAPIRegionName() *string {
	return nil
}

func (r *ContainerRegistryNetworkEgressUsage) getEgressAPIRegionName() *string {
	return nil
}

func (r *ComputeVPNGatewayNetworkEgressUsage) getEgressAPIRegionName() *string {
	return strPtr(r.Region)
}

func (r *ComputeExternalVPNGatewayNetworkEgressUsage) getEgressAPIRegionName() *string {
	return strPtr(r.Region)
}

// ****** getEgressAPIRegionName funcs ******

func (r *StorageBucketNetworkEgressUsage) getEgressAPIServiceName() *string {
	return strPtr("Cloud Storage")
}

func (r *ContainerRegistryNetworkEgressUsage) getEgressAPIServiceName() *string {
	return strPtr("Cloud Storage")
}

func (r *ComputeVPNGatewayNetworkEgressUsage) getEgressAPIServiceName() *string {
	return strPtr("Compute Engine")
}

func (r *ComputeExternalVPNGatewayNetworkEgressUsage) getEgressAPIServiceName() *string {
	return strPtr("Compute Engine")
}
