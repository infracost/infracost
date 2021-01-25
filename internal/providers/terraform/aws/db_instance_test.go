package aws_test

import (
	"fmt"
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestDBInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_db_instance" "mysql" {
			engine               = "mysql"
			instance_class       = "db.t3.large"
		}

		resource "aws_db_instance" "mysql-allocated-storage" {
			engine               = "mysql"
			instance_class       = "db.t3.large"
			allocated_storage    = 20
		}

		resource "aws_db_instance" "mysql-multi-az" {
			engine               = "mysql"
			instance_class       = "db.t3.large"
			multi_az             = true
			allocated_storage    = 30
		}

		resource "aws_db_instance" "mysql-magnetic" {
			engine               = "mysql"
			instance_class       = "db.t3.large"
			storage_type         = "standard"
			allocated_storage    = 40
		}

		resource "aws_db_instance" "mysql-iops" {
			engine               = "mysql"
			instance_class       = "db.t3.large"
			storage_type         = "io1"
			allocated_storage    = 50
			iops                 = 500
		}`

	singleAzInstanceCheck := testutil.CostComponentCheck{
		Name:            "Database instance",
		PriceHash:       "04a2cf31c0b8bf8623b1c4bd96856d49-d2c98780d7b6e36641b521f1f8145c6f",
		HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
	}

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_db_instance.mysql",
			CostComponentChecks: []testutil.CostComponentCheck{
				singleAzInstanceCheck,
				{
					Name:             "Database storage",
					PriceHash:        "b7b7cfbe7ec1bded9a474fff7123b34f-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
		{
			Name: "aws_db_instance.mysql-allocated-storage",
			CostComponentChecks: []testutil.CostComponentCheck{
				singleAzInstanceCheck,
				{
					Name:             "Database storage",
					PriceHash:        "b7b7cfbe7ec1bded9a474fff7123b34f-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(20)),
				},
			},
		},
		{
			Name: "aws_db_instance.mysql-multi-az",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Database instance",
					PriceHash:       "6533699ad0fd39e396567de86c73917b-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Database storage",
					PriceHash:        "2ec5ef73cbd5ca537c967fff828f39fe-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(30)),
				},
			},
		},
		{
			Name: "aws_db_instance.mysql-magnetic",
			CostComponentChecks: []testutil.CostComponentCheck{
				singleAzInstanceCheck,
				{
					Name:            "Database storage",
					PriceHash:       "87a57c551b26e3c6114e5034536dd82c-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(40)),
				},
			},
		},
		{
			Name: "aws_db_instance.mysql-iops",
			CostComponentChecks: []testutil.CostComponentCheck{
				singleAzInstanceCheck,
				{
					Name:            "Database storage",
					PriceHash:       "49c604321c7ca45d46173de5bdcbe1d9-ee3dd7e4624338037ca6fea0933a662f",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(50)),
				},
				{
					Name:            "Database storage IOPS",
					PriceHash:       "feb9c53577f5beba555ef9a78d59a160-9c483347596633f8cf3ab7fdd5502b78",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(500)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestDBInstance_allEngines(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	expectedEngineResults := []struct {
		Engine        string
		InstanceClass string
		PriceHash     string
	}{
		{Engine: "aurora", InstanceClass: "db.t3.small", PriceHash: "ec7901dd6154514bda21171814872566-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "aurora-mysql", InstanceClass: "db.t3.small", PriceHash: "ec7901dd6154514bda21171814872566-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "aurora-postgresql", InstanceClass: "db.t3.large", PriceHash: "c02a325181f4b5bc43827fded2393de9-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "mariadb", InstanceClass: "db.t3.large", PriceHash: "c26e17848bf2a0594017d471892782c2-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "mysql", InstanceClass: "db.t3.large", PriceHash: "04a2cf31c0b8bf8623b1c4bd96856d49-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "postgres", InstanceClass: "db.t3.large", PriceHash: "4aed0c16438fe1bce3400ded9c81e46e-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "oracle-se", InstanceClass: "db.t3.large", PriceHash: "98e5b47d043d9bdadc20f1f524751675-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "oracle-se1", InstanceClass: "db.t3.large", PriceHash: "51a21ce3143a0014bc93b6d286a10f0e-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "oracle-se2", InstanceClass: "db.t3.large", PriceHash: "7839bf8f2edb4a8ac8cc236fc042e0c7-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "oracle-ee", InstanceClass: "db.t3.large", PriceHash: "e11ffc928ba6c26619f3b6426420b6ec-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "sqlserver-ex", InstanceClass: "db.t3.large", PriceHash: "f13c7b2b683a29ba8c512253d27c92a4-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "sqlserver-web", InstanceClass: "db.t3.large", PriceHash: "7a5ab0c93fc3b3e49672cb3a1e6d7c16-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "sqlserver-se", InstanceClass: "db.m5.xlarge", PriceHash: "24dc9e9f6ca1eec2578b2db58dd5332a-d2c98780d7b6e36641b521f1f8145c6f"},
		{Engine: "sqlserver-ee", InstanceClass: "db.m5.xlarge", PriceHash: "b117119f12e72674a8748f43d7a2a70c-d2c98780d7b6e36641b521f1f8145c6f"},
	}

	tf := ""
	resourceChecks := make([]testutil.ResourceCheck, 0, len(expectedEngineResults))
	for _, expectedEngineResult := range expectedEngineResults {
		tf += fmt.Sprintf(`
		resource "aws_db_instance" "%s" {
			engine               = "%s"
			instance_class       = "%s"
		}`, expectedEngineResult.Engine, expectedEngineResult.Engine, expectedEngineResult.InstanceClass)

		resourceChecks = append(resourceChecks, testutil.ResourceCheck{
			Name: fmt.Sprintf("aws_db_instance.%s", expectedEngineResult.Engine),
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Database instance",
					PriceHash:       expectedEngineResult.PriceHash,
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Database storage",
					PriceHash:        "b7b7cfbe7ec1bded9a474fff7123b34f-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		})
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestDBInstance_byol(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_db_instance" "oracle-se1" {
			engine               = "oracle-se1"
			instance_class       = "db.t3.large"
			license_model        = "bring-your-own-license"
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_db_instance.oracle-se1",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Database instance",
					PriceHash:       "1cebf0148b5867bc68736550bdad879c-d2c98780d7b6e36641b521f1f8145c6f",
					HourlyCostCheck: testutil.HourlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
				{
					Name:             "Database storage",
					PriceHash:        "b7b7cfbe7ec1bded9a474fff7123b34f-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}
