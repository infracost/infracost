package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newInt(i int64) *int64 {
	return &i
}

func TestRegionsUsage_Values(t *testing.T) {
	tests := []struct {
		name  string
		usage RegionsUsage
		want  []RegionUsage
	}{
		{
			name: "should return non nil values as slice",
			usage: RegionsUsage{
				UsWest1:  newInt(88),
				EuWest2:  newInt(99),
				AfSouth1: newInt(77),
			},
			want: []RegionUsage{
				{"us-west-1", 88},
				{"eu-west-2", 99},
				{"af-south-1", 77},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.usage.Values()
			assert.Equal(t, tt.want, got)
		})
	}
}
