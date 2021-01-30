package aws

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCalcACMCertificateRequests(t *testing.T) {

	oneThousandCertificateRequests := decimal.NewFromInt(1000)
	tenThousandCertificateRequests := decimal.NewFromInt(10000)
	twentyThousandCertificateRequests := decimal.NewFromInt(20000)

	var certificateTierRequests = map[string]decimal.Decimal{
		"tierOne":   decimal.Zero,
		"tierTwo":   decimal.Zero,
		"tierThree": decimal.Zero,
	}

	var tierOneRequests = map[string]decimal.Decimal{
		"tierOne": decimal.NewFromInt(1000),
		"tierTwo": decimal.Zero,
	}

	var tierTwoRequests = map[string]decimal.Decimal{
		"tierOne": decimal.NewFromInt(1000),
		"tierTwo": decimal.NewFromInt(9000),
	}

	var tierThreeRequests = map[string]decimal.Decimal{
		"tierOne":   decimal.NewFromInt(1000),
		"tierTwo":   decimal.NewFromInt(10000),
		"tierThree": decimal.NewFromInt(9000),
	}

	tests := []struct {
		requests          decimal.Decimal
		inputTierRequests map[string]decimal.Decimal
		expected          map[string]decimal.Decimal
	}{
		{requests: oneThousandCertificateRequests, inputTierRequests: certificateTierRequests, expected: tierOneRequests},
		{requests: tenThousandCertificateRequests, inputTierRequests: certificateTierRequests, expected: tierTwoRequests},
		{requests: twentyThousandCertificateRequests, inputTierRequests: certificateTierRequests, expected: tierThreeRequests},
	}

	for _, test := range tests {
		actual := calculateCertificateRequests(test.requests, test.inputTierRequests)

		if test.requests == oneThousandCertificateRequests {
			assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
		}

		if test.requests == tenThousandCertificateRequests {
			assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
			assert.Equal(t, test.expected["tierTwo"], actual["tierTwo"])
		}

		if test.requests == twentyThousandCertificateRequests {
			assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
			assert.Equal(t, test.expected["tierTwo"], actual["tierTwo"])
			assert.Equal(t, test.expected["tierThree"], actual["tierThree"])
		}
	}

}
