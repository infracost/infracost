    provider "aws" {
      region                      = "us-east-1" # <<<<< Try changing this to eu-west-1 to compare the costs
  skip_credentials_validation = true
  skip_requesting_account_id  = true
  access_key                  = "mock_access_key"
  secret_key                  = "mock_secret_key"
    }

      	resource "aws_docdb_cluster" "my_cluster" {
		cluster_identifier      = "my-docdb-cluster"
		engine                  = "docdb"
		master_username         = "foo"
		master_password         = "mustbeeightchars"
		backup_retention_period = 5
		preferred_backup_window = "07:00-09:00"
		skip_final_snapshot     = true
	  }