package usage

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCalculateTierBuckets(t *testing.T) {
	oneTierBucket := decimal.NewFromInt(10000)
	twoTierBuckets := decimal.NewFromInt(15000)
	threeTierBuckets := decimal.NewFromInt(50000)
	fourTierBuckets := decimal.NewFromInt(500000)

	oneTierLimits := []int{5000}
	twoTierLimits := []int{1000, 10000}
	threeTierLimits := []int{1000, 10000, 100000}
	fourTierLimits := []int{1000, 10000, 100000, 1000000}

	var oneTierResult = []decimal.Decimal{decimal.NewFromInt(5000), decimal.NewFromInt(5000)}

	var twoTierResult = []decimal.Decimal{decimal.NewFromInt(1000), decimal.NewFromInt(10000), decimal.NewFromInt(4000)}

	var threeTierResult = []decimal.Decimal{decimal.NewFromInt(1000), decimal.NewFromInt(10000), decimal.NewFromInt(39000)}

	var fourTierResult = []decimal.Decimal{decimal.NewFromInt(1000), decimal.NewFromInt(10000), decimal.NewFromInt(100000), decimal.NewFromInt(389000)}

	tests := []struct {
		requests        decimal.Decimal
		inputTierLimits []int
		expected        []decimal.Decimal
	}{
		{requests: oneTierBucket, inputTierLimits: oneTierLimits, expected: oneTierResult},
		{requests: twoTierBuckets, inputTierLimits: twoTierLimits, expected: twoTierResult},
		{requests: threeTierBuckets, inputTierLimits: threeTierLimits, expected: threeTierResult},
		{requests: fourTierBuckets, inputTierLimits: fourTierLimits, expected: fourTierResult},
	}

	for _, test := range tests {
		actual := CalculateTierBuckets(test.requests, test.inputTierLimits)

		if test.requests == twoTierBuckets {
			assert.Equal(t, test.expected[0], actual[0])
			assert.Equal(t, test.expected[1], actual[1])
		}

		if test.requests == threeTierBuckets {
			assert.Equal(t, test.expected[0], actual[0])
			assert.Equal(t, test.expected[1], actual[1])
			assert.Equal(t, test.expected[2], actual[2])
		}

		if test.requests == fourTierBuckets {
			assert.Equal(t, test.expected[0], actual[0])
			assert.Equal(t, test.expected[1], actual[1])
			assert.Equal(t, test.expected[2], actual[2])
			assert.Equal(t, test.expected[3], actual[3])
		}
	}
}
