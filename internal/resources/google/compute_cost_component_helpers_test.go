package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getSustainedUseDiscount(t *testing.T) {
	tests := []struct {
		name  string
		hours float64
		rates sudRates
		want  float64
	}{
		{
			name:  "0 hours 20% discount",
			hours: 0,
			rates: sudRate20,
			want:  0,
		},
		{
			name:  "100 hours 20% discount",
			hours: 100,
			rates: sudRate20,
			want:  0,
		},
		{
			name:  "300 hours 20% discount",
			hours: 300,
			rates: sudRate20,
			want:  0.05,
		},
		{
			name:  "730 hours 20% discount",
			hours: 730,
			rates: sudRate20,
			want:  0.2,
		},
		{
			name:  "0 hours 30% discount",
			hours: 0,
			rates: sudRate30,
			want:  0,
		},
		{
			name:  "100 hours 30% discount",
			hours: 100,
			rates: sudRate30,
			want:  0,
		},
		{
			name:  "300 hours 30% discount",
			hours: 300,
			rates: sudRate30,
			want:  0.08,
		},
		{
			name:  "730 hours 30% discount",
			hours: 730,
			rates: sudRate30,
			want:  0.3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSustainedUseDiscount(tt.hours, tt.rates)
			assert.Equal(t, tt.want, got)
		})
	}
}
