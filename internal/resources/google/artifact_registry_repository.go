package google

import (
	"fmt"
	"regexp"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

var (
	artifactRegSvcName = strPtr("Artifact Registry")

	continentDefault      = "Intercontinental (Excl Oceania)"
	continentNorthAmerica = "North America"
	continentSouthAmerica = "South America"
	continentEurope       = "Europe"
	// GCP misspells AsiaPacific without the i. The below is not a typo!
	continentApac    = "AsiaPacfic"
	continentOceania = "Oceania"

	artifactGlobalEgressContinents = map[string]struct{}{
		continentOceania:      {},
		continentSouthAmerica: {},
	}
	artifactRegMultiRegionNames = map[string]struct{}{
		"us":     {},
		"europe": {},
		"asia":   {},
	}

	regionSep = regexp.MustCompile(`[_\-]`)
)

// ArtifactRegistryRepository struct represents a GCP artifact Registry. Artifact registry is essentially
// a next generation version of google's container registry. It allows users to store container images and language
// packages in the GCP.
//
// Pricing for Artifact Registry is based on storage amounts and data transfer.
//
// Resource information: https://cloud.google.com/artifact-registry
// Pricing information: https://cloud.google.com/artifact-registry/pricing
type ArtifactRegistryRepository struct {
	Address   string
	Region    string
	Continent string

	// StorageGB represents a usage cost that defines the amount of gb the artifact registry uses on a per monthly basis.
	StorageGB *float64 `infracost_usage:"storage_gb"`
	// MonthlyEgressDataTransferGB represents a complex usage cost that defines data transfer to different regions in the
	// google cloud infra. This does not include outbound internet egress (e.g. downloading artifact data to a local machine).
	MonthlyEgressDataTransferGB *RegionsUsage `infracost_usage:"monthly_egress_data_transfer_gb"`
}

// artifactRegistryRepositoryUsageSchema defines a list which represents the usage schema of ArtifactRegistryRepository.
func (r *ArtifactRegistryRepository) CoreType() string {
	return "ArtifactRegistryRepository"
}

func (r *ArtifactRegistryRepository) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
		{
			Key: "monthly_egress_data_transfer_gb",
			DefaultValue: &usage.ResourceUsage{
				Name:  "monthly_egress_data_transfer_gb",
				Items: RegionUsageSchema,
			},
			ValueType: schema.SubResourceUsage,
		},
	}
}

// PopulateUsage parses the u schema.UsageData into the ArtifactRegistryRepository.
// It uses the `infracost_usage` struct tags to populate data into the ArtifactRegistryRepository.
func (r *ArtifactRegistryRepository) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ArtifactRegistryRepository struct. It returns ArtifactRegistryRepository
// as a schema.Resource with two main cost components: storage costs & egress costs.
//
// Storage costs:
//
//	priced at $0.10 a month after artifact registry usage is > 0.5 GB. We ignore the free tier as there
//	is no way to currently tell if other artifact registry resources have gone beyond this free usage tier.
//
// Network costs:
//  1. free within the same region
//  2. free from multi-region to a region within the same continent, e.g. europe -> europe-west1
//  3. $0.01 when between different regions in North America continent
//  4. $0.02 when between different regions in Europe continent
//  5. $0.05 when between different regions in AsiaPacific continent
//  6. $0.15 when between any region and Oceania continent
//  7. $0.08 for all other intercontinental data transfer
//
// This method is called after the resource is initialised by an IaC provider. See providers folder for more information.
func (r *ArtifactRegistryRepository) BuildResource() *schema.Resource {
	r.Continent = continentName(r.Region)

	costComponents := []*schema.CostComponent{
		r.storageCostComponent(),
	}

	if r.MonthlyEgressDataTransferGB != nil {
		costComponents = append(costComponents, r.internalEgressComponents()...)
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *ArtifactRegistryRepository) internalEgressComponents() []*schema.CostComponent {
	filters := r.toEgressFilters()
	components := make([]*schema.CostComponent, 0, len(filters))
	for _, v := range filters {
		components = append(components, &schema.CostComponent{
			Name:            v.name,
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: v.value,
			ProductFilter: &schema.ProductFilter{
				VendorName:    vendorName,
				Region:        v.region,
				Service:       artifactRegSvcName,
				ProductFamily: strPtr("Network"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: v.desc},
					{Key: "resourceGroup", Value: strPtr("InterregionEgress")},
				},
			},
			PriceFilter: &schema.PriceFilter{
				PurchaseOption: strPtr("OnDemand"),
			},
			UsageBased: true,
		})
	}

	return components
}

func (r *ArtifactRegistryRepository) storageCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: floatPtrToDecimalPtr(r.StorageGB),
		ProductFilter: &schema.ProductFilter{
			VendorName:    vendorName,
			Service:       artifactRegSvcName,
			ProductFamily: strPtr("ApplicationServices"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("Artifact Registry Storage")},
				{Key: "resourceGroup", Value: strPtr("Storage")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("OnDemand"),
			// we ignore the free tier pricing and start at the paid pricing which is at 0.5.
			StartUsageAmount: strPtr("0.5"),
		},
		UsageBased: true,
	}
}

type artifactRegistryEgressFilters struct {
	name   string
	desc   *string
	region *string
	value  *decimal.Decimal
}

// toEgressFilters returns ArtifactRegistryRepository.MonthlyEgressDataTransferGB as a slice of filters to be used to build
// a set of cost components. One artifactRegistryEgressFilters is used by one schema.CostComponent.
//
// It uses reflection to loop over the ArtifactRegistryRepository.MonthlyEgressDataTransferGB, checking if a value has been
// given at the given region. If it has we also make sure that the region usage isn't free. See isEgressFree method for more
// details. toEgressFilters also bunches costs into continent to continent filters, as this is how they are displayed on
// the pricing page. This means that if the user specifies egress to two regions in the same continent, e.g. europe-west1 & europe-west2
// from europe-north1. Then these will be aggregated under a "Europe to Europe" artifactRegistryEgressFilters and in turn
// a schema.CostComponent.
func (r *ArtifactRegistryRepository) toEgressFilters() []artifactRegistryEgressFilters {
	if r.MonthlyEgressDataTransferGB == nil {
		return nil
	}

	values := r.MonthlyEgressDataTransferGB.Values()

	data := make([]artifactRegistryEgressFilters, 0, len(values))
	transferMap := make(map[string]int)

	for _, region := range values {
		if region.Value == 0 {
			continue
		}

		if r.isEgressFree(region.Key) {
			continue
		}

		continent := continentName(region.Key)
		value := decimal.NewFromFloat(region.Value)

		// check if the user has specified multiple regions that are based in the same continent.
		// We want to bunch these cost components into a single value output.
		name := r.egressComponentName(continent)
		if x, ok := transferMap[name]; ok {
			value = data[x].value.Add(value)
			data[x].value = &value
			continue
		}

		data = append(data, artifactRegistryEgressFilters{
			name:   name,
			desc:   r.egressDescriptionFilter(continent),
			region: r.egressRegionFilter(continent),
			value:  &value,
		})
		transferMap[name] = len(data) - 1
	}

	return data
}

func (r *ArtifactRegistryRepository) isEgressFree(region string) bool {
	// data moving within the same region is free
	if r.Region == region {
		return true
	}

	// data moving from multi-region artifact repository to a region located in the same continent as the multi-region
	// artifact repository is free.
	if _, ok := artifactRegMultiRegionNames[r.Region]; ok {
		p := regionSep.Split(region, -1)

		if p[0] == r.Region {
			return true
		}
	}

	return false
}

func (r *ArtifactRegistryRepository) egressDescriptionFilter(continent string) *string {
	if continent == continentOceania {
		return strPtr("Artifact Registry Network Inter Region Egress Intercontinental to/from Oceania")
	}

	if r.Continent == continentSouthAmerica || continent == continentSouthAmerica {
		return strPtr("Artifact Registry Network Inter Region Egress Intercontinental (Excl Oceania)")
	}

	if r.Continent == continent {
		return strPtr(fmt.Sprintf("Artifact Registry Network Inter Region Egress %s to %s", r.Continent, continent))
	}

	return strPtr("Artifact Registry Network Inter Region Egress Intercontinental (Excl Oceania)")
}

func (r *ArtifactRegistryRepository) egressComponentName(continent string) string {
	if continent == continentOceania || r.Continent == continentOceania {
		return "Data egress from/to Oceania"
	}

	if r.Continent == continentSouthAmerica || continent == continentSouthAmerica {
		return "Intercontinental (Excl Oceania)"
	}

	// replace the gcp continent naming with the correctly spelled continent
	// for the cli output.
	from := r.Continent
	if from == continentApac {
		from = "AsiaPacific"
	}

	to := continent
	if to == continentApac {
		to = "AsiaPacific"
	}

	if continent == r.Continent {
		return fmt.Sprintf("Data egress %s to %s", from, to)
	}

	return "Intercontinental (Excl Oceania)"
}

func (r *ArtifactRegistryRepository) egressRegionFilter(continent string) *string {
	if _, ok := artifactGlobalEgressContinents[r.Continent]; ok {
		return strPtr("global")
	}

	if _, ok := artifactGlobalEgressContinents[continent]; ok {
		return strPtr("global")
	}

	if r.Continent != continent {
		return strPtr("global")
	}

	return strPtr(r.Region)
}

func continentName(region string) string {
	p := regionSep.Split(region, -1)
	if len(p) == 0 {
		return continentNorthAmerica
	}

	switch p[0] {
	case "us", "northamerica":
		return continentNorthAmerica
	case "europe":
		return continentEurope
	case "asia":
		return continentApac
	case "southamerica":
		return continentSouthAmerica
	case "australia":
		return continentOceania
	}

	return continentDefault
}
