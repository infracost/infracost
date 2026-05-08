package aws

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateStorageGBSeconds(t *testing.T) {
	gbSeconds := decimal.NewFromInt(1000000)

	tests := []struct {
		name        string
		storageSize int64
		wantSign    string // "positive", "zero", "negative"
	}{
		{
			name:        "storage size below free tier (128 MB) should produce zero, not negative",
			storageSize: 128,
			wantSign:    "zero",
		},
		{
			name:        "storage size equal to free tier (512 MB) should produce zero",
			storageSize: 512,
			wantSign:    "zero",
		},
		{
			name:        "storage size above free tier (1024 MB) should produce positive result",
			storageSize: 1024,
			wantSign:    "positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateStorageGBSeconds(decimal.NewFromInt(tc.storageSize), gbSeconds)
			switch tc.wantSign {
			case "zero":
				assert.True(t, result.IsZero(), "expected zero but got %s", result)
			case "positive":
				assert.True(t, result.IsPositive(), "expected positive but got %s", result)
			case "negative":
				assert.True(t, result.IsNegative(), "expected negative but got %s", result)
			}
		})
	}
}
