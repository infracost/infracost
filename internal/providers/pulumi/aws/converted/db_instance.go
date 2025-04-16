package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDBInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_db_instance",
		RFunc:           NewDBInstance,
		ReferenceAttributes: []string{"replicateSourceDb"},
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

	iops := d.Get("iops").Float()
	defaultStorageType := "gp2"
	if iops > 0 {
		defaultStorageType = "io1"
	}

	storageType := d.GetStringOrDefault("storageType", defaultStorageType)
	r := &aws.DBInstance{
		Address:                              d.Address,
		Region:                               d.Get("region").String(),
		InstanceClass:                        d.Get("instanceClass").String(),
		Engine:                               engine,
		Version:                              d.Get("engineVersion").String(),
		MultiAZ:                              d.Get("multiAz").Bool(),
		LicenseModel:                         d.Get("licenseModel").String(),
		BackupRetentionPeriod:                d.Get("backupRetentionPeriod").Int(),
		IOPS:                                 iops,
		StorageType:                          storageType,
		IOOptimized:                          false, // IO Optimized isn't supported by terraform yet
		PerformanceInsightsEnabled:           piEnabled,
		PerformanceInsightsLongTermRetention: piLongTerm,
	}

	if !d.IsEmpty("allocated_storage") {
		r.AllocatedStorageGB = floatPtr(d.Get("allocatedStorage").Float())
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
