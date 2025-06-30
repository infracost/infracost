package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDBInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_db_instance",
		RFunc:               NewDBInstance,
		ReferenceAttributes: []string{"replicate_source_db"},
	}
}

func NewDBInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	piEnabled := d.Get("performanceInsightsEnabled").Bool()
	piLongTerm := piEnabled && d.Get("performanceInsightsRetentionPeriod").Int() > 7
	engine := d.Get("engine").String()

	replicateSourceDBs := d.References("replicateSourceDb")
	if len(replicateSourceDBs) > 0 {
		if !replicateSourceDBs[0].IsEmpty("engine") {
			engine = replicateSourceDBs[0].Get("engine").String()
		}
	}

	r := &aws.DBInstance{
		Address:                              d.Address,
		Region:                               d.Get("region").String(),
		InstanceClass:                        d.Get("instanceClass").String(),
		Engine:                               engine,
		MultiAZ:                              d.Get("multiAz").Bool(),
		LicenseModel:                         d.Get("licenseModel").String(),
		BackupRetentionPeriod:                d.Get("backupRetentionPeriod").Int(),
		IOPS:                                 d.Get("iops").Float(),
		StorageType:                          d.Get("storageType").String(),
		PerformanceInsightsEnabled:           piEnabled,
		PerformanceInsightsLongTermRetention: piLongTerm,
	}

	if !d.IsEmpty("allocatedStorage") {
		r.AllocatedStorageGB = floatPtr(d.Get("allocatedStorage").Float())
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
