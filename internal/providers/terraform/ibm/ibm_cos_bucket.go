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

	r := &ibm.IbmCosBucket{
		Address:            d.Address,
		Region:             d.Get("region").String(),
		Location:           location,
		LocationIdentifier: locationIdentifier,
		StorageClass:       storage_class,
		Archive:            archive_enabled,
		ArchiveType:        archive_type,
	}

	r.PopulateUsage(u)

	configuration := make(map[string]any)
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
