package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getDBInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_db_instance",
		RFunc: NewDbInstance,
	}
}
func NewDbInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DbInstance{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String()), InstanceClass: strPtr(d.Get("instance_class").String()), Engine: strPtr(d.Get("engine").String()), MultiAz: boolPtr(d.Get("multi_az").Bool()), LicenseModel: strPtr(d.Get("license_model").String())}
	if !d.IsEmpty("backup_retention_period") {
		r.BackupRetentionPeriod = strPtr(d.Get("backup_retention_period").String())
	}
	if !d.IsEmpty("storage_type") {
		r.StorageType = strPtr(d.Get("storage_type").String())
	}
	if !d.IsEmpty("iops") {
		r.Iops = floatPtr(d.Get("iops").Float())
	}
	if !d.IsEmpty("allocated_storage") {
		r.AllocatedStorage = floatPtr(d.Get("allocated_storage").Float())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
