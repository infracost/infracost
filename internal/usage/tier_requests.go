package usage

import (
	"strconv"

	"github.com/shopspring/decimal"
)

// TODO: can we possibly refactor this function so it handles the following cases.
//   I think the function can always returns the same number of items as in the tiers
//   array so the caller can use `if tier[x] > 0` to add that cost component.
//   Not sure if another parameter is required for the "over x" bit, or if tierBuckets is a better name...
// a) 150 requests with tiers [first 10, over 10] should return [10, 140]
// b) 150 requests with tiers [first 10, next 90, over 100] should return [10, 90, 50]
// c) 99 requests with tiers [first 10, next 90, over 100] should return [10, 89, 0]
// d) 150 requests with tiers [first 10, next 90] should return [10, 90] as there is no "over x" tier
func CalculateTierRequests(requestCount decimal.Decimal, tierLimits []int) map[string]decimal.Decimal {
	requestTierMap := map[string]decimal.Decimal{}

	for limit := range tierLimits {
		tier := decimal.NewFromInt(int64(tierLimits[limit]))

		if requestCount.GreaterThan(tier) && limit+1 < len(tierLimits) {
			requestTierMap[strconv.Itoa(limit+1)] = tier
		} else {
			requestTierMap[strconv.Itoa(limit+1)] = requestCount
		}

		if requestCount.GreaterThanOrEqual(tier) {
			requestCount = requestCount.Sub(tier)
		} else {
			requestCount = requestCount.Sub(requestCount)
		}
	}

	return requestTierMap
}
