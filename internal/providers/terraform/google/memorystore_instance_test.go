package google_test

import (
	"testing"

	googleprovider "github.com/infracost/infracost/internal/providers/terraform/google"
	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	googleresources "github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func TestMemorystoreInstance(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTests(t, "memorystore_instance_test")
}

func TestNewMemorystoreInstanceDefaultsAndUnknownCounts(t *testing.T) {
	tests := []struct {
		name             string
		values           string
		expectedNodeType string
		expectedNodes    *int64
		expectedAOFGB    *float64
		expectNil        bool
	}{
		{
			name:             "uses API defaults",
			values:           `{"location":"us-central1","shard_count":3}`,
			expectedNodeType: "HIGHMEM_MEDIUM",
			expectedNodes:    int64TestPtr(3),
			expectedAOFGB:    float64TestPtr(39),
		},
		{
			name:      "does not default an unknown node type",
			values:    `{"location":"us-central1","shard_count":3,"node_type":null}`,
			expectNil: true,
		},
		{
			name:             "preserves unknown replica count",
			values:           `{"location":"us-central1","shard_count":3,"replica_count":null}`,
			expectedNodeType: "HIGHMEM_MEDIUM",
			expectedAOFGB:    float64TestPtr(39),
		},
		{
			name:             "preserves unknown shard count",
			values:           `{"location":"us-central1","shard_count":null,"replica_count":1}`,
			expectedNodeType: "HIGHMEM_MEDIUM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := schema.NewResourceData(
				"google_memorystore_instance",
				"google",
				"google_memorystore_instance.test",
				nil,
				gjson.Parse(tt.values),
			)

			coreResource := googleprovider.NewMemorystoreInstance(data)
			if tt.expectNil {
				if coreResource != nil {
					t.Fatal("expected an unknown node type to return no resource")
				}
				return
			}

			resource, ok := coreResource.(*googleresources.MemorystoreInstance)
			if !ok {
				t.Fatal("expected a MemorystoreInstance resource")
			}
			if resource.NodeType != tt.expectedNodeType {
				t.Fatalf("expected node type %s, got %s", tt.expectedNodeType, resource.NodeType)
			}
			assertInt64PtrEqual(t, tt.expectedNodes, resource.NodeCount)
			assertFloat64PtrEqual(t, tt.expectedAOFGB, resource.AOFProvisionedGB)
		})
	}
}

func int64TestPtr(v int64) *int64 {
	return &v
}

func float64TestPtr(v float64) *float64 {
	return &v
}

func assertInt64PtrEqual(t *testing.T, expected, actual *int64) {
	t.Helper()
	if expected == nil && actual == nil {
		return
	}
	if expected == nil || actual == nil || *expected != *actual {
		t.Fatalf("expected node count %v, got %v", expected, actual)
	}
}

func assertFloat64PtrEqual(t *testing.T, expected, actual *float64) {
	t.Helper()
	if expected == nil && actual == nil {
		return
	}
	if expected == nil || actual == nil || *expected != *actual {
		t.Fatalf("expected AOF capacity %v, got %v", expected, actual)
	}
}
