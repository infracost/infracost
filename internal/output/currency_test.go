package output_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

// This tests a conversion to a currency with a non-$ symbol.  Use Panamaniam Balboas because they are pegged to USD so
// the golden file test shouldn't break with currency flucations.  Hopefully.
func TestCurrencyPBAGoldenFile(t *testing.T) {
	t.Parallel()
	tftest.GoldenFileResourceTestsWithOpts(t, "currency_pab_test", &tftest.GoldenFileOptions{Currency: "PAB"})
}

// This tests a conversion to a non USD currency that uses $ as it's symbol (we append the currency code for these).
// Use Bahamian dollars because they are pegged to USD so the golden file test shouldn't break with currency flucations.
func TestCurrencyBSDGoldenFile(t *testing.T) {
	t.Parallel()
	tftest.GoldenFileResourceTestsWithOpts(t, "currency_bsd_test", &tftest.GoldenFileOptions{Currency: "BSD"})
}
