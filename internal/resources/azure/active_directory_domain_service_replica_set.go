package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type ActiveDirectoryDomainServiceReplicaSet struct {
	Address            string
	Region             string
	DomainServiceIDSKU string
}

func (r *ActiveDirectoryDomainServiceReplicaSet) CoreType() string {
	return "ActiveDirectoryDomainServiceReplicaSet"
}

func (r *ActiveDirectoryDomainServiceReplicaSet) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *ActiveDirectoryDomainServiceReplicaSet) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ActiveDirectoryDomainServiceReplicaSet) BuildResource() *schema.Resource {
	region := r.Region

	costComponents := activeDirectoryDomainServiceCostComponents("Active directory domain service replica set", region, r.DomainServiceIDSKU)

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}
