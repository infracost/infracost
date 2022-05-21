package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	log "github.com/sirupsen/logrus"
)

type ReservedCacheNodesOfferingsParams struct {
	Region        string
	CacheNodeType string
	Duration      string
	OfferingType  string
}

func elasticacheNewClient(ctx context.Context, region string) (*elasticache.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	return elasticache.NewFromConfig(cfg), nil
}

func ElasticacheIsValidReservedOfferingType(ctx context.Context, params ReservedCacheNodesOfferingsParams) (bool, error) {
	ec, err := elasticacheNewClient(ctx, params.Region)
	if err != nil {
		return false, err
	}
	log.Debugf("Querying AWS Elasticache API: DescribeReservedCacheNodesOfferings (region: %s, cacheNodeType: %s, duration: %s, offeringType: %s", params.Region, params.CacheNodeType, params.Duration, params.OfferingType)
	result, err := ec.DescribeReservedCacheNodesOfferings(ctx, &elasticache.DescribeReservedCacheNodesOfferingsInput{
		CacheNodeType: strPtr(params.CacheNodeType),
		Duration:      strPtr(params.Duration),
		OfferingType:  strPtr(params.OfferingType),
	})
	if err != nil {
		return false, err
	}
	if len(result.ReservedCacheNodesOfferings) > 0 {
		return true, nil
	}
	return false, nil
}
