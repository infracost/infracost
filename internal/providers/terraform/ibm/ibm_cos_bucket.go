package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getIbmCosBucketRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "ibm_cos_bucket",
		RFunc:               newIbmCosBucket,
		ReferenceAttributes: []string{"resource_instance_id"},
	}
}

func getLocation(d *schema.ResourceData) (string, string) {
	const regionLocation string = "region_location"
	const singleSiteLocation string = "single_site_location"
	const crossRegionLocation string = "cross_region_location"

	if d.Get(regionLocation).Type != gjson.Null {
		return d.Get(regionLocation).String(), regionLocation
	}
	if d.Get(singleSiteLocation).Type != gjson.Null {
		return d.Get(singleSiteLocation).String(), singleSiteLocation
	}
	if d.Get(crossRegionLocation).Type != gjson.Null {
		return d.Get(crossRegionLocation).String(), crossRegionLocation
	}
	return "", ""
}

func getPlan(d *schema.ResourceData, storageClass string) string {
	// default to the standard plan
	var plan string = "standard"

	// There is only one storage class (Active) available for the One Rate plan.
	if storageClass == "onerate_active" {
		plan = "cos-one-rate-plan"
	}
	// if the reference to the parent resource (the cos instance) can
	// be resolved, then use the plan set in the parent
	cosResourceRef := d.References("resource_instance_id")

	if len(cosResourceRef) > 0 {
		cosResource := cosResourceRef[0]
		plan = cosResource.Get("plan").String()
	}

	return plan
}

func newIbmCosBucket(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	location, locationIdentifier := getLocation(d)
	storage_class := d.Get("storage_class").String()
	archive_rule := d.Get("archive_rule")
	archive_enabled := false
	archive_type := ""
	if archive_rule.IsArray() {
		archive_rule_array := archive_rule.Array()
		for _, rule := range archive_rule_array {
			enabled := rule.Get("enable").Bool()
			if enabled {
				archive_enabled = true
				archive_type = rule.Get("type").String()
			}
		}
	}

	plan := getPlan(d, storage_class)

	r := &ibm.IbmCosBucket{
		Address:            d.Address,
		Location:           location,
		LocationIdentifier: locationIdentifier,
		StorageClass:       storage_class,
		Archive:            archive_enabled,
		ArchiveType:        archive_type,
		Plan:               plan,
	}

	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["plan"] = plan
	configuration["location"] = location
	configuration["locationIdentifier"] = locationIdentifier
	configuration["storage_class"] = storage_class
	configuration["archive_enabled"] = archive_enabled
	if archive_enabled {
		configuration["archive_type"] = archive_type
	}

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
