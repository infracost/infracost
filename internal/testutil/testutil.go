package testutil

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/schema"
)

var update = flag.Bool("update", false, "update .golden files")

type CostCheckFunc func(*testing.T, *schema.CostComponent)

type ResourceCheck struct {
	Name                string
	SkipCheck           bool
	CostComponentChecks []CostComponentCheck
	SubResourceChecks   []ResourceCheck
}

type CostComponentCheck struct {
	Name             string
	PriceHash        string
	SkipCheck        bool
	HourlyCostCheck  CostCheckFunc
	MonthlyCostCheck CostCheckFunc
}

func HourlyPriceMultiplierCheck(multiplier decimal.Decimal) CostCheckFunc {
	return func(t *testing.T, c *schema.CostComponent) {
		assert.Equal(t, formatAmount(c.Price().Mul(multiplier)), formatCost(c.HourlyCost), fmt.Sprintf("unexpected hourly cost for %s", c.Name))
	}
}

func MonthlyPriceMultiplierCheck(multiplier decimal.Decimal) CostCheckFunc {
	return func(t *testing.T, c *schema.CostComponent) {
		assert.Equal(t, formatAmount(c.Price().Mul(multiplier)), formatCost(c.MonthlyCost), fmt.Sprintf("unexpected monthly cost for %s", c.Name))
	}
}

func NilMonthlyCostCheck() CostCheckFunc {
	return func(t *testing.T, c *schema.CostComponent) {
		assert.Nil(t, c.MonthlyCost, fmt.Sprintf("unexpected monthly cost for %s", c.Name))
	}
}

func TestResources(t *testing.T, resources []*schema.Resource, checks []ResourceCheck) {
	foundResources := make(map[*schema.Resource]bool)

	for _, check := range checks {
		found, r := findResource(resources, check.Name)
		assert.True(t, found, fmt.Sprintf("resource %s not found", check.Name))
		if !found {
			continue
		}

		foundResources[r] = true

		if check.SkipCheck {
			continue
		}

		TestCostComponents(t, r.CostComponents, check.CostComponentChecks)
		TestResources(t, r.SubResources, check.SubResourceChecks)
	}

	for _, r := range resources {
		if r.NoPrice {
			continue
		}

		m, ok := foundResources[r]
		assert.True(t, ok && m, fmt.Sprintf("unexpected resource %s", r.Name))
	}
}

func TestCostComponents(t *testing.T, costComponents []*schema.CostComponent, checks []CostComponentCheck) {
	foundCostComponents := make(map[*schema.CostComponent]bool)

	for _, check := range checks {
		found, c := findCostComponent(costComponents, check.Name)
		assert.True(t, found, fmt.Sprintf("cost component %s not found", check.Name))
		if !found {
			continue
		}

		foundCostComponents[c] = true

		if check.SkipCheck {
			continue
		}

		assert.Equal(t, check.PriceHash, c.PriceHash(), fmt.Sprintf("unexpected price hash for %s", c.Name))

		if check.HourlyCostCheck != nil {
			check.HourlyCostCheck(t, c)
		}

		if check.MonthlyCostCheck != nil {
			check.MonthlyCostCheck(t, c)
		}
	}

	for _, c := range costComponents {
		m, ok := foundCostComponents[c]
		assert.True(t, ok && m, fmt.Sprintf("unexpected cost component %s", c.Name))
	}
}

func findResource(resources []*schema.Resource, name string) (bool, *schema.Resource) {
	for _, resource := range resources {
		if resource.Name == name {
			return true, resource
		}
	}

	return false, nil
}

func findCostComponent(costComponents []*schema.CostComponent, name string) (bool, *schema.CostComponent) {
	for _, costComponent := range costComponents {
		if costComponent.Name == name {
			return true, costComponent
		}
	}

	return false, nil
}

func formatAmount(d decimal.Decimal) string {
	f, _ := d.Float64()
	return fmt.Sprintf("%.4f", f)
}

func formatCost(d *decimal.Decimal) string {
	if d == nil {
		return "-"
	}
	return formatAmount(*d)
}

func AssertGoldenFile(t *testing.T, goldenFilePath string, actual []byte) {
	// Load the snapshot result
	expected := []byte("")
	if _, err := os.Stat(goldenFilePath); err == nil || !os.IsNotExist(err) {
		// golden file exists, load the data
		expected, err = os.ReadFile(goldenFilePath)
		assert.NoError(t, err)
	}

	if !bytes.Equal(expected, actual) {
		if *update {
			// create the golden file dir if needed
			goldenFileDir := filepath.Dir(goldenFilePath)
			if _, err := os.Stat(goldenFileDir); err != nil {
				if os.IsNotExist(err) {
					_ = os.MkdirAll(goldenFileDir, 0755)
				}
			}

			err := os.WriteFile(goldenFilePath, actual, 0600)
			assert.NoError(t, err)
			t.Logf(fmt.Sprintf("Wrote golden file %s", goldenFilePath))
		} else {
			// Generate the diff and error message.  We don't call assert.Equal because it escapes
			// newlines (\n) and the output looks terrible.
			diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(string(expected)),
				B:        difflib.SplitLines(string(actual)),
				FromFile: "Expected",
				FromDate: "",
				ToFile:   "Actual",
				ToDate:   "",
				Context:  1,
			})

			t.Errorf(fmt.Sprintf("\nOutput does not match golden file: \n\n%s\n", diff))
		}
	}
}

type ErrorOnAnyWriter struct {
	t *testing.T
}

func (e ErrorOnAnyWriter) Write(data []byte) (n int, err error) {
	e.t.Errorf("Unexpected log write.  To capture logs remove t.Parallel and use GoldenFileOptions { CaptureLogs = true }: %s", data)
	return io.Discard.Write(data)
}

func ConfigureTestToFailOnLogs(t *testing.T, runCtx *config.RunContext) {
	runCtx.Config.LogLevel = "warn"
	runCtx.Config.SetLogDisableTimestamps(true)
	runCtx.Config.SetLogWriter(io.MultiWriter(os.Stderr, ErrorOnAnyWriter{t}))
	runCtx.Config.DisableReportCaller()

	err := logging.ConfigureBaseLogger(runCtx.Config)
	require.Nil(t, err)
}

func ConfigureTestToCaptureLogs(t *testing.T, runCtx *config.RunContext) *bytes.Buffer {
	logBuf := bytes.NewBuffer([]byte{})
	runCtx.Config.LogLevel = "warn"
	runCtx.Config.SetLogDisableTimestamps(true)
	runCtx.Config.SetLogWriter(io.MultiWriter(os.Stderr, logBuf))
	runCtx.Config.DisableReportCaller()

	err := logging.ConfigureBaseLogger(runCtx.Config)
	require.Nil(t, err)
	return logBuf
}

// From https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6
func toSnakeCase(s string) string {
	var res = make([]rune, 0, len(s))
	var p = '_'
	for i, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			res = append(res, '_')
		} else if unicode.IsUpper(r) && i > 0 {
			if unicode.IsLetter(p) && !unicode.IsUpper(p) || unicode.IsDigit(p) {
				res = append(res, '_', unicode.ToLower(r))
			} else {
				res = append(res, unicode.ToLower(r))
			}
		} else {
			res = append(res, unicode.ToLower(r))
		}

		p = r
	}
	return string(res)
}

func CalcGoldenFileTestdataDirName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("Couldn't determine currentFunctionName")
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		panic("Couldn't determine currentFunctionName")
	}

	camelCaseName := fn.Name()[strings.LastIndex(fn.Name(), ".")+1:] // slice to get everything after the last .
	if !strings.HasPrefix(camelCaseName, "Test") {
		panic(fmt.Sprintf("Please don't use this method outside of tests.  Found name: %v", camelCaseName))
	}
	return toSnakeCase(camelCaseName[4:])
}
