package putest

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/providers/pulumi"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/infracost/infracost/internal/usage"
)

// GoldenFileResourceTests runs a test for a Pulumi resource using golden files
func GoldenFileResourceTests(t *testing.T, testName string) {
	t.Helper()

	runCtx, err := config.NewRunContextFromEnv(context.Background())
	assert.NoError(t, err)

	// Load the Pulumi preview JSON file
	previewJSONPath := filepath.Join("testdata", testName, testName+".json")
	_, err = os.Stat(previewJSONPath)
	require.NoError(t, err, "Pulumi preview JSON file does not exist: %s", previewJSONPath)

	// Load the usage data, if any
	var usageData schema.UsageMap
	usageFilePath := filepath.Join("testdata", testName, testName+".usage.yml")
	if _, err := os.Stat(usageFilePath); err == nil || !os.IsNotExist(err) {
		// usage file exists, load the data
		usageFile, err := usage.LoadUsageFile(usageFilePath)
		require.NoError(t, err)
		usageData = usageFile.ToUsageDataMap()
	}

	// Create project context
	projectCtx := config.NewProjectContext(runCtx, &config.Project{
		Path: previewJSONPath,
		Name: testName,
	}, nil)

	// Create Pulumi provider
	provider := pulumi.NewPreviewJSONProvider(projectCtx, false)

	// Load resources
	projects, err := provider.LoadResources(usageData)
	require.NoError(t, err)
	require.Len(t, projects, 1)

	// Run cost calculations
	for _, project := range projects {
		project.BuildResources(usageData)
	}

	// Run price fetcher
	pf := prices.NewPriceFetcher(runCtx, true)
	for _, project := range projects {
		err = pf.PopulatePrices(project)
		require.NoError(t, err)
		schema.CalculateCosts(project)
	}

	// Generate the output
	r, err := output.ToOutputFormat(runCtx.Config, projects)
	require.NoError(t, err)
	r.Currency = runCtx.Config.Currency

	// Convert to table
	opts := output.Options{
		ShowSkipped: true,
		NoColor:     true,
		Fields:      runCtx.Config.Fields,
	}

	actual, err := output.ToTable(r, opts)
	require.NoError(t, err)

	// Check against golden file
	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

// ResourceTests runs tests for a given Pulumi JSON and checks resources against expected values
func ResourceTests(t *testing.T, previewJSON string, usage schema.UsageMap, checks []testutil.ResourceCheck) {
	t.Helper()

	// Create temp file with Pulumi JSON
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "preview.json")
	err := os.WriteFile(jsonPath, []byte(previewJSON), 0600)
	require.NoError(t, err)

	// Create contexts
	runCtx, err := config.NewRunContextFromEnv(context.Background())
	require.NoError(t, err)

	projectCtx := config.NewProjectContext(runCtx, &config.Project{
		Path: jsonPath,
		Name: t.Name(),
	}, nil)

	// Create provider
	provider := pulumi.NewPreviewJSONProvider(projectCtx, false)

	// Load resources
	projects, err := provider.LoadResources(usage)
	require.NoError(t, err)
	require.Len(t, projects, 1)

	// Build resources
	projects[0].BuildResources(usage)

	// Run cost calculations
	pf := prices.NewPriceFetcher(runCtx, true)
	err = pf.PopulatePrices(projects[0])
	require.NoError(t, err)
	schema.CalculateCosts(projects[0])

	// Test resources
	testutil.TestResources(t, projects[0].Resources, checks)
}