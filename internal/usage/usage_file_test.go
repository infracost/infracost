package usage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/usage"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestUsageFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.EnsurePluginsInstalled()
	tftest.GoldenFileUsageSyncTest(t, "usage_file")
}

func TestUsageFileNoUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.EnsurePluginsInstalled()
	tftest.GoldenFileUsageSyncTest(t, "usage_file_no_usage")
}

func TestUsageFileEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.EnsurePluginsInstalled()
	tftest.GoldenFileUsageSyncTest(t, "usage_file_empty")
}

// This should really be in schema.usage_data_test but I need to put it here so I can use LoadUsageFileFromString without
// getting import cycles.
func TestUsageDataEmpty(t *testing.T) {
	usageFile, err := usage.LoadUsageFileFromString(
		`
version: 0.1
resource_usage:
  some.resource:
    number: 0
    string: string
    object:
      k: 1
    array[0]:
      el: 2
    empty_string: ""
    empty_array[0]:
    empty:
    nested:
      number: 0
      string: string
      object:
        k: 1,
      array[0]:
        el: 2
      empty_string": ""
      empty_array[0]:
      empty:
`)
	assert.NoError(t, err)

	u := usageFile.ToUsageDataMap().Get("some.resource")

	tests := []struct {
		key  string
		want bool
	}{
		{key: "missing", want: true},
		{key: "number", want: false},
		{key: "string", want: false},
		{key: "object", want: false},
		{key: "array[0]", want: false},
		{key: "empty_string", want: true},
		{key: "empty_array[0]", want: true},
		{key: "empty", want: true},

		// nested keys don't work in usage, so these are all expected to return true
		{key: "nested.missing", want: true},
		{key: "nested.number", want: true},
		{key: "nested.string", want: true},
		{key: "nested.object", want: true},
		{key: "nested.array[0]", want: true},
		{key: "nested.empty_string", want: true},
		{key: "nested.empty_array[0]", want: true},
		{key: "nested.empty", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.want, u.IsEmpty(tt.key))
		})
	}

}
