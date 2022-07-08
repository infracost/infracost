package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestAzureRMCosmosDBCassandraKeyspace(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTests(t, "cosmosdb_cassandra_keyspace_test")
}

func TestHCLAzureRMCosmosDBCassandraKeyspace(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileHCLResourceTestsWithOpts(t, "cosmosdb_cassandra_keyspace_test_with_blank_geo_location", tftest.DefaultGoldenFileOptions())
}
