package output_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/config"
	. "github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/usage"
)

func TestToTable(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tests := []struct {
		name        string
		projectData []byte
		usageData   []byte
		want        string
		wantErr     error
	}{
		{
			name: "should build valid table output",
			projectData: []byte(`
provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_dx_connection" "my_dx_connection" {
  bandwidth = "1Gbps"
  location  = "EqDC2"
  name      = "Test"
}

resource "aws_dx_connection" "my_dx_connection_usage" {
  bandwidth = "1Gbps"
  location  = "EqDC2"
  name      = "Test_Usage"
}

resource "aws_dx_connection" "my_dx_connection_usage_backwards_compat" {
  bandwidth = "1Gbps"
  location  = "EqDC2"
  name      = "Test_Usage_Backwards"
}
`),
			usageData: []byte(`
version: 0.1
resource_usage:
  aws_dx_connection.my_dx_connection_usage:
    monthly_outbound_region_to_dx_location_gb: 100
    monthly_outbound_from_region_to_dx_connection_location:
      eu_west_1: 1000
      ap_east_1: 3000
      does_not_exist: 6000
    dx_virtual_interface_type: private
    dx_connection_type: dedicated
  aws_dx_connection.my_dx_connection_usage_backwards_compat:
    monthly_outbound_region_to_dx_location_gb: 200
    dx_virtual_interface_type: private
    dx_connection_type: dedicated
`),
			want: `
 Name                                                       Monthly Qty  Unit   Monthly Cost 
                                                                                             
 aws_dx_connection.my_dx_connection                                                          
 └─ DX connection                                                   730  hours       $219.00 
                                                                                             
 aws_dx_connection.my_dx_connection_usage                                                    
 ├─ DX connection                                                   730  hours       $219.00 
 ├─ Outbound data transfer (from ap-east-1, to EqDC2)             3,000  GB          $270.00 
 └─ Outbound data transfer (from eu-west-1, to EqDC2)             1,000  GB           $28.20 
                                                                                             
 aws_dx_connection.my_dx_connection_usage_backwards_compat                                   
 ├─ DX connection                                                   730  hours       $219.00 
 └─ Outbound data transfer (from us-east-1, to EqDC2)               200  GB            $4.00 
                                                                                             
 OVERALL TOTAL                                                                       $959.20 `,
		},
		{
			name: "should skip zero value cost component",
			projectData: []byte(`
provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_dx_connection" "my_dx_connection" {
  bandwidth = "1Gbps"
  location  = "EqDC2"
  name      = "Test"
}

resource "aws_dx_connection" "should_not_show_ap_east_1" {
  bandwidth = "1Gbps"
  location  = "EqDC2"
  name      = "Test_Usage"
}
`),
			usageData: []byte(`
version: 0.1
resource_usage:
  aws_dx_connection.should_not_show_ap_east_1:
    monthly_outbound_from_region_to_dx_connection_location:
      eu_west_1: 1000
      ap_east_1: 0
`),
			want: `
 Name                                                  Monthly Qty  Unit   Monthly Cost 
                                                                                        
 aws_dx_connection.my_dx_connection                                                     
 └─ DX connection                                              730  hours       $219.00 
                                                                                        
 aws_dx_connection.should_not_show_ap_east_1                                            
 ├─ DX connection                                              730  hours       $219.00 
 └─ Outbound data transfer (from eu-west-1, to EqDC2)        1,000  GB           $28.20 
                                                                                        
 OVERALL TOTAL                                                                  $466.20 `,
		},
		{
			name: "should skip resource entirely",
			projectData: []byte(`provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_lambda_function" "should_show" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "should_not_show" {
  function_name = "lambda_function_name"
  role          = "arn:aws:lambda:us-east-1:account-id:resource-id"
  handler       = "exports.test"
  runtime       = "nodejs12.x"
  memory_size   = 512
}
`),
			usageData: []byte(`
version: 0.1
resource_usage:
  aws_lambda_function.should_show:
    monthly_requests: 100000
    request_duration_ms: 350

  aws_lambda_function.should_not_show:
    monthly_requests: 0
`),
			want: `
 Name                             Monthly Qty  Unit         Monthly Cost 
                                                                         
 aws_lambda_function.should_show                                         
 ├─ Requests                              0.1  1M requests         $0.02 
 └─ Duration                            4,375  GB-seconds          $0.07 
                                                                         
 OVERALL TOTAL                                                     $0.09 `,
		},
		{
			name: "should remove sub resources",
			projectData: []byte(`

provider "aws" {
  region                      = "us-east-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  skip_get_ec2_platforms      = true
  skip_region_validation      = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
}

resource "aws_s3_bucket" "usage" {
  bucket = "bucket2_withUsage"
}
`),
			usageData: []byte(`
version: 0.1
resource_usage:
  aws_s3_bucket.usage:
    standard:
      storage_gb: 10000
      monthly_select_data_scanned_gb: 0
`),
			want: `
 Name                                             Monthly Qty  Unit                    Monthly Cost 
                                                                                                    
 aws_s3_bucket.usage                                                                                
 └─ Standard                                                                                        
    ├─ Storage                                         10,000  GB                           $230.00 
    ├─ PUT, COPY, POST, LIST requests       Monthly cost depends on usage: $0.005 per 1k requests   
    ├─ GET, SELECT, and all other requests  Monthly cost depends on usage: $0.0004 per 1k requests  
    └─ Select data returned                 Monthly cost depends on usage: $0.0007 per GB           
                                                                                                    
 OVERALL TOTAL                                                                              $230.00 
----------------------------------
To estimate usage-based resources use --usage-file, see https://infracost.io/usage-file`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCtx, err := config.NewRunContextFromEnv(context.Background())

			assert.NoError(t, err)
			tfProject := tftest.TerraformProject{
				Files: []tftest.File{
					{
						Path:     "main.tf",
						Contents: string(tt.projectData),
					},
				},
			}

			u, err := usage.ParseYaml(tt.usageData)
			assert.NoError(t, err)

			projects, err := tftest.RunCostCalculations(t, runCtx, tfProject, u)
			assert.NoError(t, err)
			assert.Len(t, projects, 1)

			r := ToOutputFormat(projects)
			r.Currency = runCtx.Config.Currency
			actual, err := ToTable(r, Options{
				ShowSkipped: true,
				NoColor:     true,
				Fields:      runCtx.Config.Fields,
			})

			assert.Equal(t, tt.wantErr, err)

			endOfFirstLine := bytes.Index(actual, []byte("\n"))
			if endOfFirstLine > 0 {
				actual = actual[endOfFirstLine+1:]
			}
			assert.Equal(t, tt.want, string(actual))
		})
	}
}
