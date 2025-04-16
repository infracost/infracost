package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getSQLDatabaseInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_sql_database_instance",
		RFunc: NewSQLDatabaseInstance,
		Notes: []string{
			"Cloud SQL network, SQL Server license, 1-3 years commitments costs are not yet supported.",
		},
	}
}

func NewSQLDatabaseInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.SQLDatabaseInstance{
		Address:              d.Address,
		DatabaseVersion:      d.Get("databaseVersion").String(),
		Tier:                 d.Get("settings.0.tier").String(),
		Edition:              d.Get("settings.0.edition").String(),
		Region:               d.Get("region").String(),
		ReplicaConfiguration: d.Get("replicaConfiguration").String(),
		AvailabilityType:     d.Get("settings.0.availabilityType").String(),
		DiskType:             d.Get("settings.0.diskType").String(),
		DiskSize:             d.Get("settings.0.diskSize").Int(),
	}
	if !d.IsEmpty("settings.0.ip_configuration.0.ipv4_enabled") {
		r.UseIPV4 = d.Get("settings.0.ipConfiguration.0.ipv4Enabled").Bool()
	} else {
		r.UseIPV4 = true // Should use ipv4 if the ipv4_enabled is empty
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
