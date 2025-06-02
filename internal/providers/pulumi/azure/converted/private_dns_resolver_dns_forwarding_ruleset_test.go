package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestPrivateDnsResolverDnsForwardingRuleset(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tftest.GoldenFileResourceTests(t, "private_dns_resolver_dns_forwarding_ruleset_test")
}
