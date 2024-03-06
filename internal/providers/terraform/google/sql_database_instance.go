package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getSQLDatabaseInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_sql_database_instance",
		CoreRFunc: NewSQLDatabaseInstance,
		Notes: []string{
			"Cloud SQL network, SQL Server license, 1-3 years commitments costs are not yet supported.",
		},
	}
}

func NewSQLDatabaseInstance(d *schema.ResourceData) schema.CoreResource {
	r := &google.SQLDatabaseInstance{
		Address:              d.Address,
		DatabaseVersion:      d.Get("database_version").String(),
		Tier:                 d.Get("settings.0.tier").String(),
		Edition:              d.Get("settings.0.edition").String(),
		Region:               d.Get("region").String(),
		ReplicaConfiguration: d.Get("replica_configuration").String(),
		AvailabilityType:     d.Get("settings.0.availability_type").String(),
		DiskType:             d.Get("settings.0.disk_type").String(),
		DiskSize:             d.Get("settings.0.disk_size").Int(),
	}
	if !d.IsEmpty("settings.0.ip_configuration.0.ipv4_enabled") {
		r.UseIPV4 = d.Get("settings.0.ip_configuration.0.ipv4_enabled").Bool()
	} else {
		r.UseIPV4 = true // Should use ipv4 if the ipv4_enabled is empty
	}
	return r
}
