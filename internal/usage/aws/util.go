package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	log "github.com/sirupsen/logrus"
)

const timeMonth = time.Hour * 24 * 30

func sdkWarn(service string, usageType string, id string, err interface{}) {
	// HACK: too busy to figure out how to make logrus print to screen
	fmt.Printf("Error estimating %s %s usage for %s: %s\n", service, usageType, id, err)
	log.Warnf("Error estimating %s %s usage for %s: %s", service, usageType, id, err)
}

func sdkNewConfig(region string) (aws.Config, error) {
	return config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
}
