package ibm

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

// Database struct represents a database instance
//
// This terraform resource is opaque and can handle multiple databases, provided with the right parameters
type Database struct {
	Name     string
	Address  string
	Service  string
	Plan     string
	Location string
	Group    gjson.Result
	Flavor   string
	Disk     int64
	Memory   int64
	CPU      int64
	Members  int64
}

type DatabaseCostComponentsFunc func(*Database) []*schema.CostComponent

// PopulateUsage parses the u schema.UsageData into the Database.
// It uses the `infracost_usage` struct tags to populate data into the Database.
func (r *Database) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// DatabaseUsageSchema defines a list which represents the usage schema of Database.
var DatabaseUsageSchema = []*schema.UsageItem{}

var DatabaseCostMap map[string]DatabaseCostComponentsFunc = map[string]DatabaseCostComponentsFunc{
	"databases-for-postgresql": GetPostgresCostComponents,
	// "databases-for-etcd":
	// "databases-for-redis":
	"databases-for-elasticsearch": GetElasticSearchCostComponents,
	// "messages-for-rabbitmq":
	// "databases-for-mongodb":
	// "databases-for-mysql":
	// "databases-for-cassandra":
	// "databases-for-enterprisedb"
}

// BuildResource builds a schema.Resource from a valid Database struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *Database) BuildResource() *schema.Resource {
	costComponentsFunc, ok := DatabaseCostMap[r.Service]

	if !ok {
		return &schema.Resource{
			Name:        r.Address,
			UsageSchema: DatabaseUsageSchema,
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    DatabaseUsageSchema,
		CostComponents: costComponentsFunc(r),
	}
}
