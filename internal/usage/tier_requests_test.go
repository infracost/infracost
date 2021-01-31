package usage

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalculateTierRequests(t *testing.T) {
	twoTierRequests := decimal.NewFromInt(15000)
	threeTierRequests := decimal.NewFromInt(50000)
	fourTierRequests := decimal.NewFromInt(500000)

	twoTierLimits := []int{1000, 10000}
	threeTierLimits := []int{1000, 10000, 100000}
	fourTierLimits := []int{1000, 10000, 100000, 1000000}

	var twoTierMap = map[string]decimal.Decimal{
		"1": decimal.NewFromInt(1000),
		"2": decimal.NewFromInt(14000),
	}

	var threeTierMap = map[string]decimal.Decimal{
		"1": decimal.NewFromInt(1000),
		"2": decimal.NewFromInt(10000),
		"3": decimal.NewFromInt(39000),
	}

	var fourTierMap = map[string]decimal.Decimal{
		"1": decimal.NewFromInt(1000),
		"2": decimal.NewFromInt(10000),
		"3": decimal.NewFromInt(100000),
		"4": decimal.NewFromInt(389000),
	}

	tests := []struct {
		requests          decimal.Decimal
		inputTierRequests []int
		expected          map[string]decimal.Decimal
	}{
		{requests: twoTierRequests, inputTierRequests: twoTierLimits, expected: twoTierMap},
		{requests: threeTierRequests, inputTierRequests: threeTierLimits, expected: threeTierMap},
		{requests: fourTierRequests, inputTierRequests: fourTierLimits, expected: fourTierMap},
	}

	for _, test := range tests {
		actual := CalculateTierRequests(test.requests, test.inputTierRequests)

		if test.requests == twoTierRequests {
			assert.Equal(t, test.expected["1"], actual["1"])
			assert.Equal(t, test.expected["2"], actual["2"])
		}

		if test.requests == threeTierRequests {
			assert.Equal(t, test.expected["1"], actual["1"])
			assert.Equal(t, test.expected["2"], actual["2"])
			assert.Equal(t, test.expected["3"], actual["3"])
		}

		if test.requests == fourTierRequests {
			assert.Equal(t, test.expected["1"], actual["1"])
			assert.Equal(t, test.expected["2"], actual["2"])
			assert.Equal(t, test.expected["3"], actual["3"])
			assert.Equal(t, test.expected["4"], actual["4"])
		}
	}
}
