package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegionsUsage_Values(t *testing.T) {
	tests := []struct {
		name  string
		usage RegionsUsage
		want  []RegionUsage
	}{
		{
			name: "should return non nil values as slice",
			usage: RegionsUsage{
				USWest1:  floatPtr(88),
				EUWest2:  floatPtr(99),
				AFSouth1: floatPtr(77),
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
