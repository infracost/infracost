package aws

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalcApiRequests(t *testing.T) {
	oneMillionAPI := decimal.NewFromInt(1000000)
	oneBillionAPI := decimal.NewFromInt(1000000000)

	var apiTierRequests = map[string]decimal.Decimal{
		"tierOne":   decimal.Zero,
		"tierTwo":   decimal.Zero,
		"tierThree": decimal.Zero,
		"tierFour":  decimal.Zero,
	}

	var apiTierOneRequests = map[string]decimal.Decimal{
		"tierOne":   decimal.NewFromInt(1000000),
		"tierTwo":   decimal.Zero,
		"tierThree": decimal.Zero,
		"tierFour":  decimal.Zero,
	}

	var apiTierTwoRequests = map[string]decimal.Decimal{
		"tierOne":   decimal.NewFromInt(333000000),
		"tierTwo":   decimal.NewFromInt(667000000),
		"tierThree": decimal.Zero,
		"tierFour":  decimal.Zero,
	}

	tests := []struct {
		requests          decimal.Decimal
		inputTierRequests map[string]decimal.Decimal
		expected          map[string]decimal.Decimal
	}{
		{requests: oneMillionAPI, inputTierRequests: apiTierRequests, expected: apiTierOneRequests},
		{requests: oneBillionAPI, inputTierRequests: apiTierRequests, expected: apiTierTwoRequests},
	}

	for _, test := range tests {
		actual := calculateAPIRequests(test.requests, test.inputTierRequests)

		if test.requests == oneMillionAPI {
			assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
		}

		if test.requests == oneBillionAPI {
			assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
			assert.Equal(t, test.expected["tierTwo"], actual["tierTwo"])
		}
	}
}
