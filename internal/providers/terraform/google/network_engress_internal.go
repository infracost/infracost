package google

import (
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

type EgressResourceType int

const (
	StorageBucketEgress EgressResourceType = iota
	ContainerRegistryEgress
	ComputeVPNGateway
	ComputeExternalVPNGateway
)

type egressRegionData struct {
	gRegion             string
	apiDescription      string
	apiDescriptionRegex string
	usageKey            string
	fixedRegion         string
}

type egressRegionUsageFilterData struct {
	usageNumber int64
	usageName   string
}

func networkEgress(region string, u *schema.UsageData, resourceName, prefixName string, egressResourceType EgressResourceType) *schema.Resource {
	resource := &schema.Resource{
		Name:           resourceName,
		CostComponents: []*schema.CostComponent{},
	}

	// Same continent
	if doesEgressIncludeSameContinent(egressResourceType) {
		var quantity *decimal.Decimal
		if u != nil {
			if v, ok := u.Get("monthly_egress_data_transfer_gb").Map()["same_continent"]; ok {
				quantity = decimalPtr(decimal.NewFromInt(v.Int()))
			}
		}

		resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
			Name:            fmt.Sprintf("%s in same continent", prefixName),
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
		})
	}

	regionsData := getEgressRegionsData(prefixName, egressResourceType)
	usageFiltersData := getEgressUsageFiltersData(egressResourceType)
	egressUsageMap := map[string]gjson.Result{}
	if u != nil {
		egressUsageMap = u.Get("monthly_egress_data_transfer_gb").Map()
	}

	for _, regData := range regionsData {
		gRegion := regData.gRegion
		usageKey := regData.usageKey

		// TODO: Reformat to use tier helpers.
		var usage int64
		var used int64
		var lastEndUsageAmount int64
		if v, ok := egressUsageMap[usageKey]; ok {
			usage = v.Int()
		}

		for idx, usageFilter := range usageFiltersData {
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
				usageFilter = ""
			}
			if quantity == nil && idx > 0 {
				continue
			}
			var apiRegion *string
			if regData.fixedRegion != "" {
				apiRegion = strPtr(regData.fixedRegion)
			} else {
				apiRegion = getEgressAPIRegionName(region, egressResourceType)
			}
			var name string
			if usageName != "" {
				name = fmt.Sprintf("%v (%v)", gRegion, usageName)
			} else {
				name = fmt.Sprintf("%v", gRegion)
			}
			attributeFilters := make([]*schema.AttributeFilter, 0)
			if regData.apiDescriptionRegex != "" {
				attributeFilters = append(attributeFilters, &schema.AttributeFilter{Key: "description", ValueRegex: strPtr(regData.apiDescriptionRegex)})
			} else {
				attributeFilters = append(attributeFilters, &schema.AttributeFilter{Key: "description", Value: strPtr(regData.apiDescription)})
			}
			resource.CostComponents = append(resource.CostComponents, &schema.CostComponent{
				Name:            name,
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: quantity,
				ProductFilter: &schema.ProductFilter{
					Region:           apiRegion,
					VendorName:       strPtr("gcp"),
					Service:          getEgressAPIServiceName(egressResourceType),
					AttributeFilters: attributeFilters,
				},
				PriceFilter: &schema.PriceFilter{
					EndUsageAmount: strPtr(usageFilter),
				},
			})
		}
	}

	return resource
}

func doesEgressIncludeSameContinent(egressResourceType EgressResourceType) bool {
	switch egressResourceType {
	case ComputeExternalVPNGateway, ComputeVPNGateway:
		return false
	default:
		return true
	}
}

func getEgressRegionsData(prefixName string, egressResourceType EgressResourceType) []*egressRegionData {
	switch egressResourceType {
	case ComputeVPNGateway:
		return []*egressRegionData{
			{
				gRegion: fmt.Sprintf("%s within the same region", prefixName),
				// There is no same region option in APIs, so we always take this price in us-central1 region.
				apiDescription: "Network Inter Region Data Transfer Out from Americas to Virginia",
				usageKey:       "same_region",
				fixedRegion:    "us-central1",
			},
			{
				gRegion:        fmt.Sprintf("%s within the US or Canada", prefixName),
				apiDescription: "Network Inter Region Data Transfer Out from Americas to Montreal",
				usageKey:       "us_or_canada",
				fixedRegion:    "us-central1",
			},
			{
				gRegion:        fmt.Sprintf("%s within Europe", prefixName),
				apiDescription: "Network Inter Region Data Transfer Out from EMEA to Frankfurt",
				usageKey:       "europe",
				fixedRegion:    "europe-west1",
			},
			{
				gRegion:        fmt.Sprintf("%s within Asia", prefixName),
				apiDescription: "Network Inter Region Data Transfer Out from Japan to Seoul",
				usageKey:       "asia",
				fixedRegion:    "asia-northeast1",
			},
			{
				gRegion:        fmt.Sprintf("%s within South America", prefixName),
				apiDescription: "Network Inter Region Data Transfer Out from Sao Paulo to Sao Paulo",
				usageKey:       "south_america",
				fixedRegion:    "southamerica-east1",
			},
			{
				gRegion:        fmt.Sprintf("%s to/from Indonesia and Oceania", prefixName),
				apiDescription: "Network Inter Region Data Transfer Out from Sydney to Jakarta",
				usageKey:       "oceania",
				fixedRegion:    "australia-southeast1",
			},
			{
				gRegion:        fmt.Sprintf("%s between continents (excludes Oceania)", prefixName),
				apiDescription: "Network Inter Region Data Transfer Out from Finland to Singapore",
				usageKey:       "worldwide",
				fixedRegion:    "europe-north1",
			},
		}

	case ComputeExternalVPNGateway:
		return []*egressRegionData{
			{
				gRegion: fmt.Sprintf("%s to worldwide excluding China, Australia but including Hong Kong", prefixName),
				// There is no worldwide option in APIs, so we take a random region.
				apiDescriptionRegex: "/Vpn Internet Egress .* to Americas/",
				usageKey:            "worldwide",
			},
			{
				gRegion:             fmt.Sprintf("%s to China excluding Hong Kong", prefixName),
				apiDescriptionRegex: "/Vpn Internet Egress .* to China/",
				usageKey:            "china",
			},
			{
				gRegion:             fmt.Sprintf("%s to Australia", prefixName),
				apiDescriptionRegex: "/Vpn Internet Egress .* to Australia/",
				usageKey:            "australia",
			},
		}
	default:
		return []*egressRegionData{
			{
				gRegion:        fmt.Sprintf("%s to worldwide excluding Asia, Australia", prefixName),
				apiDescription: "Download Worldwide Destinations (excluding Asia & Australia)",
				usageKey:       "worldwide",
			},
			{
				gRegion:        fmt.Sprintf("%s to Asia excluding China, but including Hong Kong", prefixName),
				apiDescription: "Download APAC",
				usageKey:       "asia",
			},
			{
				gRegion:        fmt.Sprintf("%s to China excluding Hong Kong", prefixName),
				apiDescription: "Download China",
				usageKey:       "china",
			},
			{
				gRegion:        fmt.Sprintf("%s to Australia", prefixName),
				apiDescription: "Download Australia",
				usageKey:       "australia",
			},
		}
	}
}

func getEgressUsageFiltersData(egressResourceType EgressResourceType) []*egressRegionUsageFilterData {
	switch egressResourceType {
	case ComputeVPNGateway:
		return []*egressRegionUsageFilterData{
			{
				usageNumber: 0,
			},
		}
	default:
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
}

func getEgressAPIRegionName(region string, egressResourceType EgressResourceType) *string {
	switch egressResourceType {
	case ComputeExternalVPNGateway, ComputeVPNGateway:
		return strPtr(region)
	default:
		return nil
	}
}

func getEgressAPIServiceName(egressResourceType EgressResourceType) *string {
	switch egressResourceType {
	case ComputeExternalVPNGateway, ComputeVPNGateway:
		return strPtr("Compute Engine")
	default:
		return strPtr("Cloud Storage")
	}
}
