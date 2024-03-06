package google

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
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

func (r *StorageBucketNetworkEgressUsage) CoreType() string {
	return "StorageBucketNetworkEgressUsage"
}

func (r *StorageBucketNetworkEgressUsage) UsageSchema() []*schema.UsageItem {
	return StorageBucketNetworkEgressUsageSchema
}

func (r *StorageBucketNetworkEgressUsage) BuildResource() *schema.Resource {
	regionsData := []*egressRegionData{
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
	usageFiltersData := []*egressRegionUsageFilterData{
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
	serviceName := "Cloud Storage"

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
				{Key: "description", Value: strPtr("Network Data Transfer GCP Inter Region within Europe")},
			},
		},
		UsageBased: true,
	})

	for _, regData := range regionsData {
		usageKey := regData.usageKey
		usage := GetFloatFieldValueByUsageTag(usageKey, *r)
		newCostComponents := egressStepPricingHelper(usage, usageFiltersData, regData, "", serviceName)
		resource.CostComponents = append(resource.CostComponents, newCostComponents...)
	}

	return resource
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

func (r *ContainerRegistryNetworkEgressUsage) CoreType() string {
	return "ContainerRegistryNetworkEgressUsage"
}

func (r *ContainerRegistryNetworkEgressUsage) UsageSchema() []*schema.UsageItem {
	return ContainerRegistryNetworkEgressUsageSchema
}

func (r *ContainerRegistryNetworkEgressUsage) BuildResource() *schema.Resource {
	regionsData := []*egressRegionData{
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
	usageFiltersData := []*egressRegionUsageFilterData{
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
	serviceName := "Cloud Storage"

	resource := &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{},
	}

	// Same continent
	var quantity *decimal.Decimal
	if r.SameContinent != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.SameContinent))
	}

	var startUsage *string
	continent := regionToContinent(r.Region)
	// Northern America has three prices available to it, only the start usage 100 is applicable to us, as this is what is reflected in the
	// pricing calculator.
	if continent == "Northern America" {
		startUsage = strPtr("100")
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
				{Key: "description", Value: strPtr(fmt.Sprintf("Network Data Transfer GCP Inter Region within %s", continent))},
				{Key: "resourceGroup", Value: strPtr("InterregionEgress")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: startUsage,
		},
		UsageBased: true,
	})

	for _, regData := range regionsData {
		usageKey := regData.usageKey
		usage := GetFloatFieldValueByUsageTag(usageKey, *r)
		newCostComponents := egressStepPricingHelper(usage, usageFiltersData, regData, "", serviceName)
		resource.CostComponents = append(resource.CostComponents, newCostComponents...)
	}

	return resource
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

func (r *ComputeVPNGatewayNetworkEgressUsage) CoreType() string {
	return "ComputeVPNGatewayNetworkEgressUsage"
}

func (r *ComputeVPNGatewayNetworkEgressUsage) UsageSchema() []*schema.UsageItem {
	return ComputeVPNGatewayNetworkEgressUsageSchema
}

func (r *ComputeVPNGatewayNetworkEgressUsage) BuildResource() *schema.Resource {
	regionsData := []*egressRegionData{
		{
			gRegion: fmt.Sprintf("%s within the same region", r.PrefixName),
			// There is no same region option in APIs, so we always take this price in us-central1 region.
			apiDescription: "Network Inter Region Data Transfer Out from Americas to Virginia",
			usageKey:       "same_region",
			fixedRegion:    "us-central1",
		},
		{
			gRegion:        fmt.Sprintf("%s within the US or Canada", r.PrefixName),
			apiDescription: "Network Inter Region Data Transfer Out from Americas to Montreal",
			usageKey:       "us_or_canada",
			fixedRegion:    "us-central1",
		},
		{
			gRegion:        fmt.Sprintf("%s within Europe", r.PrefixName),
			apiDescription: "Network Inter Region Data Transfer Out from EMEA to Frankfurt",
			usageKey:       "europe",
			fixedRegion:    "europe-west1",
		},
		{
			gRegion:        fmt.Sprintf("%s within Asia", r.PrefixName),
			apiDescription: "Network Inter Region Data Transfer Out from Japan to Seoul",
			usageKey:       "asia",
			fixedRegion:    "asia-northeast1",
		},
		{
			gRegion:        fmt.Sprintf("%s within South America", r.PrefixName),
			apiDescription: "Network Inter Region Data Transfer Out from Sao Paulo to Sao Paulo",
			usageKey:       "south_america",
			fixedRegion:    "southamerica-east1",
		},
		{
			gRegion:        fmt.Sprintf("%s to/from Indonesia and Oceania", r.PrefixName),
			apiDescription: "Network Inter Region Data Transfer Out from Sydney to Jakarta",
			usageKey:       "oceania",
			fixedRegion:    "australia-southeast1",
		},
		{
			gRegion:        fmt.Sprintf("%s between continents (excludes Oceania)", r.PrefixName),
			apiDescription: "Network Inter Region Data Transfer Out from Finland to Singapore",
			usageKey:       "worldwide",
			fixedRegion:    "europe-north1",
		},
	}
	usageFiltersData := []*egressRegionUsageFilterData{
		{
			usageNumber: 0,
		},
	}
	defaultAPIRegionName := r.Region
	serviceName := "Compute Engine"

	resource := &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{},
	}

	for _, regData := range regionsData {
		usageKey := regData.usageKey
		usage := GetFloatFieldValueByUsageTag(usageKey, *r)
		newCostComponents := egressStepPricingHelper(usage, usageFiltersData, regData, defaultAPIRegionName, serviceName)
		resource.CostComponents = append(resource.CostComponents, newCostComponents...)
	}

	return resource
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

func (r *ComputeExternalVPNGatewayNetworkEgressUsage) CoreType() string {
	return "ComputeExternalVPNGatewayNetworkEgressUsage"
}

func (r *ComputeExternalVPNGatewayNetworkEgressUsage) UsageSchema() []*schema.UsageItem {
	return ComputeExternalVPNGatewayNetworkEgressUsageSchema
}

func (r *ComputeExternalVPNGatewayNetworkEgressUsage) BuildResource() *schema.Resource {
	regionsData := []*egressRegionData{
		{
			gRegion: fmt.Sprintf("%s to worldwide excluding China, Australia but including Hong Kong", r.PrefixName),
			// There is no worldwide option in APIs, so we take a random region.
			apiDescription: "Network Vpn Internet Data Transfer Out from Americas to Americas",
			usageKey:       "worldwide",
		},
		{
			gRegion:        fmt.Sprintf("%s to China excluding Hong Kong", r.PrefixName),
			apiDescription: "Network Vpn Internet Data Transfer Out from Americas to China",
			usageKey:       "china",
		},
		{
			gRegion:        fmt.Sprintf("%s to Australia", r.PrefixName),
			apiDescription: "Network Vpn Internet Data Transfer Out from Americas to Australia",
			usageKey:       "australia",
		},
	}
	usageFiltersData := []*egressRegionUsageFilterData{
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
	defaultAPIRegionName := r.Region
	serviceName := "Compute Engine"

	resource := &schema.Resource{
		Name:           r.Address,
		CostComponents: []*schema.CostComponent{},
	}

	for _, regData := range regionsData {
		usageKey := regData.usageKey
		usage := GetFloatFieldValueByUsageTag(usageKey, *r)
		newCostComponents := egressStepPricingHelper(usage, usageFiltersData, regData, defaultAPIRegionName, serviceName)
		resource.CostComponents = append(resource.CostComponents, newCostComponents...)
	}

	return resource
}

type egressRegionData struct {
	gRegion             string // gRegion is the name used in pricing pages that is more human friendly.
	apiDescription      string
	apiDescriptionRegex string
	usageKey            string
	fixedRegion         string // fixedRegion is the region used in pricing API.
}

type egressRegionUsageFilterData struct {
	usageNumber float64
	usageName   string
}

func egressStepPricingHelper(usage float64, usageFiltersData []*egressRegionUsageFilterData, regData *egressRegionData, defaultAPIRegionName, serviceName string) []*schema.CostComponent {
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
		} else if defaultAPIRegionName != "" {
			apiRegion = strPtr(defaultAPIRegionName)
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
				Service:          strPtr(serviceName),
				AttributeFilters: attributeFilters,
			},
			PriceFilter: &schema.PriceFilter{
				EndUsageAmount: strPtr(usageFilter),
			},
			UsageBased: true,
		})
	}
	return costComponents
}
