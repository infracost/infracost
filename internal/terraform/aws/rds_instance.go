package aws

import (
	"plancosts/pkg/base"

	"github.com/shopspring/decimal"
)

var multiAzMapping = base.ValueMapping{
	FromKey: "multi_az",
	ToKey:   "deploymentOption",
	ToValueFn: func(fromVal interface{}) string {
		if fromVal.(bool) {
			return "Multi-AZ"
		}
		return "Single-AZ"
	},
}

type RdsStorageIOPS struct {
	*BaseAwsPriceComponent
}

func NewRdsStorageIOPS(name string, resource *RdsInstance) *RdsStorageIOPS {
	c := &RdsStorageIOPS{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "month"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonRDS"},
		{Key: "productFamily", Value: "Provisioned IOPS"},
		{Key: "deploymentOption", Value: "Single-AZ"},
	}

	c.valueMappings = []base.ValueMapping{multiAzMapping}

	return c
}

func (c *RdsStorageIOPS) HourlyCost() decimal.Decimal {
	hourlyCost := c.BaseAwsPriceComponent.HourlyCost()
	size := decimal.NewFromInt(int64(0))
	if c.AwsResource().RawValues()["iops"] != nil {
		size = decimal.NewFromFloat(c.AwsResource().RawValues()["iops"].(float64))
	}
	return hourlyCost.Mul(size)
}

type RdsStorageGB struct {
	*BaseAwsPriceComponent
}

func NewRdsStorageGB(name string, resource *RdsInstance) *RdsStorageGB {
	c := &RdsStorageGB{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "month"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonRDS"},
		{Key: "productFamily", Value: "Database Storage"},
		{Key: "deploymentOption", Value: "Single-AZ"},
		{Key: "volumeType", Value: "General Purpose"},
	}

	c.valueMappings = []base.ValueMapping{
		{
			FromKey: "storage_type",
			ToKey:   "volumeType",
			ToValueFn: func(fromVal interface{}) string {
				switch fromVal.(string) {
				case "standard":
					return "Magnetic"
				case "io1":
					return "Provisioned IOPS"
				}
				return "General Purpose"
			},
		},
		multiAzMapping,
	}

	return c
}

func (c *RdsStorageGB) HourlyCost() decimal.Decimal {
	hourlyCost := c.BaseAwsPriceComponent.HourlyCost()
	size := decimal.NewFromInt(int64(0))
	if c.AwsResource().RawValues()["allocated_storage"] != nil {
		size = decimal.NewFromFloat(c.AwsResource().RawValues()["allocated_storage"].(float64))
	}
	if c.AwsResource().RawValues()["max_allocated_storage"] != nil {
		size = decimal.NewFromFloat(c.AwsResource().RawValues()["max_allocated_storage"].(float64))
	}
	return hourlyCost.Mul(size)
}

type RdsInstanceHours struct {
	*BaseAwsPriceComponent
}

func NewRdsInstanceHours(name string, resource *RdsInstance) *RdsInstanceHours {
	c := &RdsInstanceHours{
		NewBaseAwsPriceComponent(name, resource.BaseAwsResource, "hour"),
	}

	c.defaultFilters = []base.Filter{
		{Key: "servicecode", Value: "AmazonRDS"},
		{Key: "productFamily", Value: "Database Instance"},
		{Key: "deploymentOption", Value: "Single-AZ"},
	}

	c.valueMappings = []base.ValueMapping{
		{FromKey: "instance_class", ToKey: "instanceType"},
		{
			FromKey: "engine",
			ToKey:   "databaseEngine",
			ToValueFn: func(fromVal interface{}) string {
				switch fromVal.(string) {
				case "postgresql":
					return "PostgreSQL"
				case "mysql":
					return "MySQL"
				case "mariadb":
					return "MariaDB"
				case "aurora", "aurora-mysql":
					return "Aurora MySQL"
				case "aurora-postgresql":
					return "Aurora PostgreSQL"
				case "oracle-se", "oracle-se1", "oracle-se2", "oracle-ee":
					return "Oracle"
				case "sqlserver-ex", "sqlserver-web", "sqlserver-se", "sqlserver-ee":
					return "SQL Server"
				}
				return ""
			},
		},
		{
			FromKey: "engine",
			ToKey:   "databaseEdition",
			ToValueFn: func(fromVal interface{}) string {
				switch fromVal.(string) {
				case "oracle-se", "sqlserver-se":
					return "Standard"
				case "oracle-se1":
					return "Standard One"
				case "oracle-se2":
					return "Standard 2"
				case "oracle-ee", "sqlserver-ee":
					return "Enterprise"
				case "sqlserver-ex":
					return "Express"
				case "sqlserver-web":
					return "Web"
				}
				return ""
			},
		},
		multiAzMapping,
	}

	return c
}

type RdsInstance struct {
	*BaseAwsResource
}

func NewRdsInstance(address string, region string, rawValues map[string]interface{}) *RdsInstance {
	r := &RdsInstance{
		NewBaseAwsResource(address, region, rawValues),
	}
	priceComponents := []base.PriceComponent{
		NewRdsInstanceHours("Instance hours", r),
		NewRdsStorageGB("GB", r),
	}
	if r.RawValues()["storage_type"] == "io1" {
		priceComponents = append(priceComponents, NewRdsStorageIOPS("IOPS", r))
	}
	r.BaseAwsResource.priceComponents = priceComponents
	return r
}
