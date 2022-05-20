package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	log "github.com/sirupsen/logrus"
)

var reservedElasticacheReservedCacheNodesOfferingMapping = map[string]string{
	"heavy_utilization":  "Heavy Utilization",
	"medium_utilization": "Medium Utilization",
	"light_utilization":  "Light Utilization",
}

type ReservedCacheNodesOfferingsParams struct {
	region        *string
	cacheNodeType *string
	duration      *string
	offeringType  *string
}

func elasticacheNewClient(ctx context.Context, region string) (*elasticache.Client, error) {
	cfg, err := getConfig(ctx, region)
	if err != nil {
		return nil, err
	}
	return elasticache.NewFromConfig(cfg), nil
}

func ElasticacheIsValidReservedOfferingType(ctx context.Context, params ReservedCacheNodesOfferingsParams) (bool, error) {
	ec, err := elasticacheNewClient(ctx, *params.region)
	if err != nil {
		return false, err
	}
	log.Debugf("Querying AWS Elasticache API: DescribeReservedCacheNodesOfferings (region: %s, cacheNodeType: %s, duration: %s, offeringType: %s", *params.region, *params.cacheNodeType, *params.duration, *params.offeringType)
	offeringType := reservedElasticacheReservedCacheNodesOfferingMapping[*params.offeringType]
	result, err := ec.DescribeReservedCacheNodesOfferings(ctx, &elasticache.DescribeReservedCacheNodesOfferingsInput{
		CacheNodeType: params.cacheNodeType,
		Duration:      params.duration,
		OfferingType:  strPtr(offeringType),
	})
	if err != nil {
		return false, err
	}
	if len(result.ReservedCacheNodesOfferings) > 0 {
		return true, nil
	}
	return false, nil
}
