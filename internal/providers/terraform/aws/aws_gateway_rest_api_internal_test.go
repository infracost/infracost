package aws

import (
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestCalcApiRequests(t *testing.T)  {
    oneMillionApi := decimal.NewFromInt(1000000)
    oneBillionApi := decimal.NewFromInt( 1000000000)

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

    tests := []struct{
        requests decimal.Decimal
        inputTierRequests map[string]decimal.Decimal
        expected map[string]decimal.Decimal
    }{
        {requests: oneMillionApi,  inputTierRequests: apiTierRequests, expected: apiTierOneRequests},
        {requests: oneBillionApi, inputTierRequests: apiTierRequests, expected: apiTierTwoRequests},
    }

    for _, test := range tests {
        actual := calculateApiRequests(test.requests, test.inputTierRequests)

        if test.requests == oneMillionApi {
            assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
        }

        if test.requests == oneBillionApi {
            assert.Equal(t, test.expected["tierOne"], actual["tierOne"])
            assert.Equal(t, test.expected["tierTwo"], actual["tierTwo"])
        }
    }
}