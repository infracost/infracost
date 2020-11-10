package output

import (
	"fmt"

	"github.com/infracost/infracost/internal/providers/terraform"
	"github.com/infracost/infracost/internal/schema"
)

func skippedResourcesMessage(resources []*schema.Resource, showDetails bool) string {
	summary := schema.GenerateResourceSummary(resources)
	if summary.TotalUnsupported == 0 {
		return ""
	}

	supportedTypeCount := 0
	for rType := range summary.UnsupportedCounts {
		if terraform.HasSupportedProvider(rType) {
			supportedTypeCount++
		}
	}

	message := fmt.Sprintf("%d resource types couldn't be estimated as Infracost doesn't support them yet (https://www.infracost.io/docs/supported_resources)", supportedTypeCount)
	if supportedTypeCount == 1 {
		message = "1 resource type couldn't be estimated as Infracost doesn't support it yet (https://www.infracost.io/docs/supported_resources)"
	}

	if showDetails {
		message += ".\n"
	} else {
		message += ", re-run with --show-skipped to see the list.\n"
	}

	message += "We're continually adding new resources, please email hello@infracost.io if you'd like us to prioritize your list."

	if showDetails {
		for rType, count := range summary.UnsupportedCounts {
			if terraform.HasSupportedProvider(rType) {
				message += fmt.Sprintf("\n%d x %s", count, rType)
			}
		}
	}

	return message
}
