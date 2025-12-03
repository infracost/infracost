//nolint:deadcode,unused
package aws

import (
	"time"

	"github.com/infracost/infracost/internal/logging"
)

const timeMonth = time.Hour * 24 * 30

func sdkWarn(service string, usageType string, id string, err any) {
	logging.Logger.Warn().Msgf("Error estimating %s %s usage for %s: %s", service, usageType, id, err)
}

func intPtr(i int64) *int64 {
	return &i
}

func int32Ptr(i int32) *int32 {
	return &i
}

func strPtr(s string) *string {
	return &s
}
