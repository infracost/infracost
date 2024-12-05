package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestRDSClusterGoldenFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTestsWithOpts(t, "rds_cluster_test", &tftest.GoldenFileOptions{
		CaptureLogs: true,
	})
}

func TestRDSClusterChinaGoldenFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTestsWithOpts(t, "rds_cluster_china_test", &tftest.GoldenFileOptions{
		CaptureLogs: true,
		Currency:    "CNY",
	})
}
