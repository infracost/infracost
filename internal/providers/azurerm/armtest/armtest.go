package armtest

import (
	"bytes"
	"context"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infracost/infracost/internal/output"
	"github.com/infracost/infracost/internal/providers/azurerm"
	"github.com/infracost/infracost/internal/usage"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

type Project struct {
	Files []File
}

type File struct {
	Path     string
	Contents string
}

func ResourceTests(t *testing.T, contents string, usage map[string]*schema.UsageData, checks []testutil.ResourceCheck) {
	project := Project{
		Files: []File{
			{
				Path:     "what_if.json",
				Contents: contents,
			},
		},
	}

	t.Run("AzureRM", func(t *testing.T) {
		resourceTests(t, "azurerm", project, usage, checks)
	})
}

func resourceTests(t *testing.T, pName string, project Project, usage map[string]*schema.UsageData, checks []testutil.ResourceCheck) {
	t.Helper()

	runCtx, err := config.NewRunContextFromEnv(context.Background())
	assert.NoError(t, err)

	projects := loadResources(t, pName, project, runCtx, usage)

	projects, err = RunCostCalculations(runCtx, projects)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)

	testutil.TestResources(t, projects[0].Resources, checks)
}

type GoldenFileOptions = struct {
	Currency    string
	CaptureLogs bool
}

func DefaultGoldenFileOptions() *GoldenFileOptions {
	return &GoldenFileOptions{
		Currency:    "USD",
		CaptureLogs: false,
	}
}

func GoldenFileResourceTests(t *testing.T, testName string) {
	GoldenFileResourceTestsWithOpts(t, testName, DefaultGoldenFileOptions())
}

func GoldenFileResourceTestsWithOpts(t *testing.T, testName string, options *GoldenFileOptions) {
	t.Run("AzureRM", func(t *testing.T) {
		goldenFileResourceTestWithOpts(t, "azurerm", testName, options)
	})
}

func goldenFileResourceTestWithOpts(t *testing.T, pName string, testName string, options *GoldenFileOptions) {
	t.Helper()

	runCtx, err := config.NewRunContextFromEnv(context.Background())

	var logBuf *bytes.Buffer
	if options != nil && options.CaptureLogs {
		logBuf = testutil.ConfigureTestToCaptureLogs(t, runCtx)
	} else {
		// testutil.ConfigureTestToFailOnLogs(t, runCtx)
	}

	if options != nil && options.Currency != "" {
		runCtx.Config.Currency = options.Currency
	}

	require.NoError(t, err)

	// Load the terraform projects
	projectData, err := os.ReadFile(filepath.Join("../testdata", testName, "what_if.json"))
	require.NoError(t, err)
	project := Project{
		Files: []File{
			{
				Path:     "what_if.json",
				Contents: string(projectData),
			},
		},
	}

	// Load the usage data, if any.
	var usageData map[string]*schema.UsageData
	usageFilePath := filepath.Join("../testdata", testName, testName+".usage.yml")
	if _, err := os.Stat(usageFilePath); err == nil || !os.IsNotExist(err) {
		// usage file exists, load the data
		usageFile, err := usage.LoadUsageFile(usageFilePath)
		require.NoError(t, err)
		usageData = usageFile.ToUsageDataMap()
	}

	projects := loadResources(t, pName, project, runCtx, usageData)

	// Generate the output
	projects, err = RunCostCalculations(runCtx, projects)
	require.NoError(t, err)

	r, err := output.ToOutputFormat(projects)
	if err != nil {
		require.NoError(t, err)
	}
	r.Currency = runCtx.Config.Currency

	opts := output.Options{
		ShowSkipped: true,
		NoColor:     true,
		Fields:      runCtx.Config.Fields,
	}

	actual, err := output.ToTable(r, opts)
	require.NoError(t, err)

	// strip the first line of output since it contains the temporary project path
	endOfFirstLine := bytes.Index(actual, []byte("\n"))
	if endOfFirstLine > 0 {
		actual = actual[endOfFirstLine+1:]
	}

	if logBuf != nil && logBuf.Len() > 0 {
		actual = append(actual, "\nLogs:\n"...)

		// need to sort the logs so they can be compared consistently
		logLines := strings.Split(logBuf.String(), "\n")
		sort.Strings(logLines)
		actual = append(actual, strings.Join(logLines, "\n")...)
	}

	goldenFilePath := filepath.Join("../testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func loadResources(t *testing.T, pName string, project Project, runCtx *config.RunContext, usageData map[string]*schema.UsageData) []*schema.Project {
	t.Helper()

	projectDir, err := CreateTestProject(t, project)
	require.NoError(t, err)

	runCtx.Config.RootPath = projectDir
	projectCtx := config.NewProjectContext(runCtx, &config.Project{
		Path: path.Join(projectDir, "what_if.json"),
	}, log.Fields{})
	provider := azurerm.NewWhatifJsonProvider(projectCtx, true)

	projects, err := provider.LoadResources(usageData)
	require.NoError(t, err)

	for _, project := range projects {
		project.BuildResources(map[string]*schema.UsageData{})
	}

	return projects
}

func RunCostCalculations(runCtx *config.RunContext, projects []*schema.Project) ([]*schema.Project, error) {
	for _, project := range projects {
		err := prices.PopulatePrices(runCtx, project)
		if err != nil {
			return projects, err
		}

		schema.CalculateCosts(project)
	}

	return projects, nil
}

func GoldenFileUsageSyncTest(t *testing.T, testName string) {
	t.Run("AzureRM", func(t *testing.T) {
		goldenFileSyncTest(t, "azurerm", testName)
	})
}

func goldenFileSyncTest(t *testing.T, pName, testName string) {
	runCtx, err := config.NewRunContextFromEnv(context.Background())
	require.NoError(t, err)

	tfProjectData, err := os.ReadFile(filepath.Join("testdata", testName, testName+".tf"))
	require.NoError(t, err)
	tfProject := Project{
		Files: []File{
			{
				Path:     "main.tf",
				Contents: string(tfProjectData),
			},
		},
	}

	projectCtx := config.NewProjectContext(runCtx, &config.Project{}, log.Fields{})

	usageFilePath := filepath.Join("testdata", testName, testName+"_existing_usage.yml")
	projects := loadResources(t, pName, tfProject, runCtx, map[string]*schema.UsageData{})

	actual := RunSyncUsage(t, projectCtx, projects, usageFilePath)
	require.NoError(t, err)

	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func RunSyncUsage(t *testing.T, projectCtx *config.ProjectContext, projects []*schema.Project, usageFilePath string) []byte {
	t.Helper()
	usageFile, err := usage.LoadUsageFile(usageFilePath)
	require.NoError(t, err)

	_, err = usage.SyncUsageData(projectCtx, usageFile, projects)
	require.NoError(t, err)

	out := filepath.Join(t.TempDir(), "actual_usage.yml")
	err = usageFile.WriteToPath(out)
	require.NoError(t, err)

	b, err := os.ReadFile(out)
	require.NoError(t, err)

	return b
}

func CreateTestProject(t *testing.T, project Project) (string, error) {
	t.Helper()
	tmpDir := t.TempDir()

	return writeToTmpDir(tmpDir, project)
}

func writeToTmpDir(tmpDir string, project Project) (string, error) {
	var err error

	for _, file := range project.Files {
		fullPath := filepath.Join(tmpDir, file.Path)
		dir := filepath.Dir(fullPath)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return tmpDir, err
			}
		}

		err = os.WriteFile(fullPath, []byte(file.Contents), 0600)
		if err != nil {
			return tmpDir, err
		}
	}

	return tmpDir, err
}
