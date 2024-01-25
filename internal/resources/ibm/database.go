package ibm

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
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

	// Databases For PostgreSQL
	// Catalog Link: https://cloud.ibm.com/catalog/services/databases-for-postgresql
	// Pricing Link: https://cloud.ibm.com/docs/databases-for-postgresql?topic=databases-for-postgresql-pricing
	RAM     *int64 `infracost_usage:"database_ram_mb"`
	Disk    *int64 `infracost_usage:"database_disk_mb"`
	Core    *int64 `infracost_usage:"database_core"`
	Members *int64 `infracost_usage:"database_members"`
}

type DatabaseCostComponentsFunc func(*Database) []*schema.CostComponent

// PopulateUsage parses the u schema.UsageData into the Database.
// It uses the `infracost_usage` struct tags to populate data into the Database.
func (r *Database) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// DatabaseUsageSchema defines a list which represents the usage schema of Database.
var DatabaseUsageSchema = []*schema.UsageItem{
	{Key: "database_ram_mb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "database_disk_mb", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "database_core", DefaultValue: 0, ValueType: schema.Int64},
	{Key: "database_members", DefaultValue: 0, ValueType: schema.Int64},
}

var DatabaseCostMap map[string]DatabaseCostComponentsFunc = map[string]DatabaseCostComponentsFunc{
	"databases-for-postgresql": GetPostgresCostComponents,
	// "databases-for-etcd":
	// "databases-for-redis":
	// "databases-for-elasticsearch":
	// "messages-for-rabbitmq":
	// "databases-for-mongodb":
	// "databases-for-mysql":
	// "databases-for-cassandra":
	// "databases-for-enterprisedb"
}

func ConvertMBtoGB(d decimal.Decimal) decimal.Decimal {
	return d.Div(decimal.NewFromInt(1024))
}

func PostgresRAMCostComponent(r *Database) *schema.CostComponent {
	var R decimal.Decimal
	if r.RAM != nil {
		R = ConvertMBtoGB(decimal.NewFromInt(*r.RAM))
	} else { // set the default
		R = decimal.NewFromInt(1)
	}
	var m decimal.Decimal
	if r.Members != nil {
		m = decimal.NewFromInt(*r.Members)
	} else { // set the default
		m = decimal.NewFromInt(2)
	}

	cost := R.Mul(m)
	
	costComponent := schema.CostComponent{
		Name:            "RAM",
		Unit:            "GB-RAM",
		MonthlyQuantity: &cost,
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr("databases-for-postgresql"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_MONTHS_RAM"),
		},
	}
	return &costComponent
}

func PostgresDiskCostComponent(r *Database) *schema.CostComponent {
	var d decimal.Decimal
	if r.Disk != nil {
		d = ConvertMBtoGB(decimal.NewFromInt(*r.Disk))
	} else { // set the default
		d = decimal.NewFromInt(5)
	}
	var m decimal.Decimal
	if r.Members != nil {
		m = decimal.NewFromInt(*r.Members)
	} else { // set the default
		m = decimal.NewFromInt(2)
	}

	cost := d.Mul(m)

	costComponent := schema.CostComponent{
		Name:            "Disk",
		Unit:            "GB-DISK",
		MonthlyQuantity: &cost,
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr("databases-for-postgresql"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("GIGABYTE_MONTHS_DISK"),
		},
	}
	return &costComponent
}

func PostgresCoreCostComponent(r *Database) *schema.CostComponent {
	var c decimal.Decimal
	if r.Core != nil {
		c = decimal.NewFromInt(*r.Core)
	} else { // set the default
		c = decimal.NewFromInt(0)
	}
	var m decimal.Decimal
	if r.Members != nil {
		m = decimal.NewFromInt(*r.Members)
	} else { // set the default
		m = decimal.NewFromInt(2)
	}

	cost := c.Mul(m)

	costComponent := schema.CostComponent{
		Name:            "Core",
		Unit:            "Virtual Processor Core",
		MonthlyQuantity: &cost,
		UnitMultiplier:  decimal.NewFromInt(1),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("ibm"),
			Region:        strPtr(r.Location),
			Service:       strPtr("databases-for-postgresql"),
			ProductFamily: strPtr("service"),
		},
		PriceFilter: &schema.PriceFilter{
			Unit: strPtr("VIRTUAL_PROCESSOR_CORES"),
		},
	}
	return &costComponent
}

func GetPostgresCostComponents(r *Database) []*schema.CostComponent {
	return []*schema.CostComponent{
		PostgresRAMCostComponent(r),
		PostgresDiskCostComponent(r),
		PostgresCoreCostComponent(r),
	}
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
