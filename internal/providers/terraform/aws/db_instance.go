package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDBInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_db_instance",
		CoreRFunc:           NewDBInstance,
		ReferenceAttributes: []string{"replicate_source_db"},
	}
}

func NewDBInstance(d *schema.ResourceData) schema.CoreResource {
	piEnabled := d.Get("performance_insights_enabled").Bool()
	piLongTerm := piEnabled && d.Get("performance_insights_retention_period").Int() > 7
	engine := d.Get("engine").String()

	replicateSourceDBs := d.References("replicate_source_db")
	if len(replicateSourceDBs) > 0 {
		if !replicateSourceDBs[0].IsEmpty("engine") {
			engine = replicateSourceDBs[0].Get("engine").String()
		}
	}

	r := &aws.DBInstance{
		Address:                              d.Address,
		Region:                               d.Get("region").String(),
		InstanceClass:                        d.Get("instance_class").String(),
		Engine:                               engine,
		MultiAZ:                              d.Get("multi_az").Bool(),
		LicenseModel:                         d.Get("license_model").String(),
		BackupRetentionPeriod:                d.Get("backup_retention_period").Int(),
		IOPS:                                 d.Get("iops").Float(),
		StorageType:                          d.Get("storage_type").String(),
		PerformanceInsightsEnabled:           piEnabled,
		PerformanceInsightsLongTermRetention: piLongTerm,
	}

	if !d.IsEmpty("allocated_storage") {
		r.AllocatedStorageGB = floatPtr(d.Get("allocated_storage").Float())
	}

	return r
}
