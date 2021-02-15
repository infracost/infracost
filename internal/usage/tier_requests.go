package usage

import (
	"github.com/shopspring/decimal"
)

// Use this method to calculate resource's tiers
// In tierLimits send only [next] tiers (first-tier used as next-tier)
// 'Over' tier calculated as the remainder of (requests - sum of requests in 'next' tiers)
//
// Examples:
// a) 150 requests (requestCount param) with tierLimits [first 10] should return [10, 140], where 140 is 'over' tier
// b) 150 requests with tiers [first 10, next 90] should return [10, 90, 50]
// c) 99 requests with tiers [first 10, next 90] should return [10, 89, 0]
// d) 100 requests with tiers [first 10, next 100, next 100] should return [10, 90, 0, 0]
//
// Method always returns array of length (len('tierLimits') + 1 (for 'over' tier))
// If you want to use it, but your resource doesn't have 'over' tier - do not use last value of returned array

func CalculateTierBuckets(requestCount decimal.Decimal, tierLimits []int) []decimal.Decimal {
	overTier := false
	tiers := make([]decimal.Decimal, 0)

	for limit := range tierLimits {
		tier := decimal.NewFromInt(int64(tierLimits[limit]))

		if requestCount.GreaterThanOrEqual(tier) {
			tiers = append(tiers, tier)
			requestCount = requestCount.Sub(tier)
			overTier = true
		} else if requestCount.LessThan(tier) {
			tiers = append(tiers, requestCount)
			requestCount = decimal.Zero
			overTier = false
		}
	}

	if overTier {
		tiers = append(tiers, requestCount)
	} else {
		tiers = append(tiers, decimal.NewFromInt(0))
	}
	return tiers
}
