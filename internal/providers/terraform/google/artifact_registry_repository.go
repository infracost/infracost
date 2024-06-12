package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getArtifactRegistryRepositoryRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_artifact_registry_repository",
		CoreRFunc: newArtifactRegistryRepository,
		GetRegion: func(defaultRegion string, d *schema.ResourceData) string {
			region := d.Get("region").String()

			zone := d.Get("zone").String()
			if zone != "" {
				region = zoneToRegion(zone)
			}

			location := d.Get("location").String()
			if location != "" {
				region = location
			}

			return region
		},
	}
}

func newArtifactRegistryRepository(d *schema.ResourceData) schema.CoreResource {
	r := &google.ArtifactRegistryRepository{
		Address: d.Address,
		Region:  d.Region,
	}

	return r
}
