package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetDBInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_db_instance",
		RFunc: NewDBInstance,
	}
}
func NewDBInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.DBInstance{Address: strPtr(d.Address), MultiAz: boolPtr(d.Get("multi_az").Bool()), InstanceClass: strPtr(d.Get("instance_class").String()), Engine: strPtr(d.Get("engine").String()), LicenseModel: strPtr(d.Get("license_model").String()), Region: strPtr(d.Get("region").String())}
	if d.Get("storage_type").Exists() && d.Get("storage_type").Type != gjson.Null {
		r.StorageType = strPtr(d.Get("storage_type").String())
	}
	if d.Get("iops").Exists() && d.Get("iops").Type != gjson.Null {
		r.Iops = floatPtr(d.Get("iops").Float())
	}
	if d.Get("allocated_storage").Exists() && d.Get("allocated_storage").Type != gjson.Null {
		r.AllocatedStorage = floatPtr(d.Get("allocated_storage").Float())
	}
	if d.Get("backup_retention_period").Exists() && d.Get("backup_retention_period").Type != gjson.Null {
		r.BackupRetentionPeriod = strPtr(d.Get("backup_retention_period").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
