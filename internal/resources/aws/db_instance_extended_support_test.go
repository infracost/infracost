package aws

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLExtendedSupportCostComponent(t *testing.T) {
	quantity := decimal.NewFromInt(2)
	region := "us-east-1"

	afterMySQL80Year1 := time.Date(2026, time.August, 2, 0, 0, 0, 0, time.UTC)
	beforeMySQL84Year1 := time.Date(2029, time.July, 31, 0, 0, 0, 0, time.UTC)
	afterMySQL84Year1 := time.Date(2029, time.August, 2, 0, 0, 0, 0, time.UTC)
	afterMySQL57Year1 := time.Date(2024, time.March, 2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name                string
		version             string
		date                time.Time
		wantNil             bool
		wantName            string
		wantUsagetypeSubstr string
	}{
		{
			name:    "mysql 8.4 is not charged after 8.0 extended support starts",
			version: "8.4",
			date:    afterMySQL80Year1,
			wantNil: true,
		},
		{
			name:    "mysql 8.4.3 is not charged after 8.0 extended support starts",
			version: "8.4.3",
			date:    afterMySQL80Year1,
			wantNil: true,
		},
		{
			name:    "mysql 8.4 is not charged just before its own extended support starts",
			version: "8.4",
			date:    beforeMySQL84Year1,
			wantNil: true,
		},
		{
			name:                "mysql 8.4 is charged after its extended support starts",
			version:             "8.4",
			date:                afterMySQL84Year1,
			wantName:            "Extended support (year 1)",
			wantUsagetypeSubstr: "ExtendedSupport:Yr1-Yr2:MySQL8.4",
		},
		{
			name:                "mysql 8.0 is charged after 8.0 extended support starts",
			version:             "8.0",
			date:                afterMySQL80Year1,
			wantName:            "Extended support (year 1)",
			wantUsagetypeSubstr: "ExtendedSupport:Yr1-Yr2:MySQL8",
		},
		{
			name:                "mysql 8.0.36 strips to pre-8.4 fallback",
			version:             "8.0.36",
			date:                afterMySQL80Year1,
			wantName:            "Extended support (year 1)",
			wantUsagetypeSubstr: "ExtendedSupport:Yr1-Yr2:MySQL8",
		},
		{
			name:                "mysql 8.1 uses pre-8.4 fallback",
			version:             "8.1",
			date:                afterMySQL80Year1,
			wantName:            "Extended support (year 1)",
			wantUsagetypeSubstr: "ExtendedSupport:Yr1-Yr2:MySQL8",
		},
		{
			name:                "mysql 8.2 uses pre-8.4 fallback",
			version:             "8.2",
			date:                afterMySQL80Year1,
			wantName:            "Extended support (year 1)",
			wantUsagetypeSubstr: "ExtendedSupport:Yr1-Yr2:MySQL8",
		},
		{
			name:                "mysql 8.3 uses pre-8.4 fallback",
			version:             "8.3",
			date:                afterMySQL80Year1,
			wantName:            "Extended support (year 1)",
			wantUsagetypeSubstr: "ExtendedSupport:Yr1-Yr2:MySQL8",
		},
		{
			name:                "mysql 5.7 still charged after its extended support starts",
			version:             "5.7",
			date:                afterMySQL57Year1,
			wantName:            "Extended support (year 1)",
			wantUsagetypeSubstr: "ExtendedSupport:Yr1-Yr2:MySQL5.7",
		},
		{
			name:                "mysql 5.7.44 strips to 5.7",
			version:             "5.7.44",
			date:                afterMySQL57Year1,
			wantName:            "Extended support (year 1)",
			wantUsagetypeSubstr: "ExtendedSupport:Yr1-Yr2:MySQL5.7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mysqlExtendedSupport.CostComponent(tt.version, region, tt.date, quantity)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)
			assert.Equal(t, tt.wantName, got.Name)
			require.NotNil(t, got.ProductFilter)
			require.NotEmpty(t, got.ProductFilter.AttributeFilters)

			var usagetype string
			for _, f := range got.ProductFilter.AttributeFilters {
				if f.Key == "usagetype" && f.ValueRegex != nil {
					usagetype = *f.ValueRegex
					break
				}
			}
			assert.Contains(t, usagetype, tt.wantUsagetypeSubstr)
		})
	}
}
