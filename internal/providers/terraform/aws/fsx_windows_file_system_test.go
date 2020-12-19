package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestFSXWindowsFS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	  resource "aws_vpc" "example" {
		cidr_block = "10.0.0.0/16"
	  }
	  
	  resource "aws_subnet" "example" {
		vpc_id     = aws_vpc.example.id
		cidr_block = "10.0.1.0/24"
	  
		tags = {
		  Name = "Main"
		}
	  }
	  
	  resource "aws_fsx_windows_file_system" "example" {
		storage_capacity    = 300
		subnet_ids          = [aws_subnet.example.id]
		throughput_capacity = 1024
		deployment_type = "MULTI_AZ_1"
 		storage_type = "HDD"
	  
		self_managed_active_directory {
		  dns_ips     = ["10.0.0.111", "10.0.0.222"]
		  domain_name = "corp.example.com"
		  password    = "avoid-plaintext-passwords"
		  username    = "Admin"
		}
	  
	  }`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_fsx_windows_file_system.example",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Throughput capacity (Multi-AZ)",
					PriceHash:       "73198d547f5abe1f94b63bced66fb5b7-8191dc82cee9b89717087e447a40abbd",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1024)),
				},
				{
					Name:             "Storage (Multi-AZ - HDD)",
					PriceHash:        "b71f728d01d84b4b0feb2ab70c58c58d-ee3dd7e4624338037ca6fea0933a662f",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(300)),
				},
				{
					Name:      "Backup Storage (Multi-AZ)",
					PriceHash: "a34f7034871d3dff9109cb011c2c97bb-ee3dd7e4624338037ca6fea0933a662f",
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, resourceChecks)
}
