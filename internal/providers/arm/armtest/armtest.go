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
	"github.com/infracost/infracost/internal/usage"

	"github.com/stretchr/testify/assert"

	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/prices"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/arm"
)

type ARMProject struct {
	Files []File
}

type File struct {
	Path     string
	Contents string
}

func ResourceTests(t *testing.T, contents string, usage schema.UsageMap, checks []testutil.ResourceCheck) {
	project := ARMProject{
		Files: []File{
			{
				Path:     "main.json",
				Contents: contents,
			},
		},
	}

	ResourceTestsForARMProject(t, project, usage, checks)
}

func ResourceTestsForARMProject(t *testing.T, armProject ARMProject, usage schema.UsageMap, checks []testutil.ResourceCheck, ctxOptions ...func(ctx *config.RunContext)) {
	t.Run("ARM", func(t *testing.T) {
		resourceTestsForARMProject(t, armProject, usage, checks, ctxOptions...)
	})
}

func resourceTestsForARMProject(t *testing.T, armProject ARMProject, usage schema.UsageMap, checks []testutil.ResourceCheck, ctxOptions ...func(ctx *config.RunContext)) {
	t.Helper()

	runCtx, err := config.NewRunContextFromEnv(context.Background())
	assert.NoError(t, err)

	for _, ctxOption := range ctxOptions {
		ctxOption(runCtx)
	}

	projects := loadResources(t, armProject, runCtx, usage)

	projects, err = RunCostCalculations(runCtx, projects)
	assert.NoError(t, err)
	assert.Len(t, projects, 1)

	testutil.TestResources(t, projects[0].Resources, checks)
}

type GoldenFileOptions = struct {
	Currency    string
	CaptureLogs bool
	IgnoreCLI   bool
	LogLevel    *string
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

func GoldenFileResourceTestsWithOpts(t *testing.T, testName string, options *GoldenFileOptions, ctxOptions ...func(ctx *config.RunContext)) {
	t.Run("ARM", func(t *testing.T) {
		goldenFileResourceTestWithOpts(t, testName, options, ctxOptions...)
	})
}

func goldenFileResourceTestWithOpts(t *testing.T, testName string, options *GoldenFileOptions, ctxOptions ...func(ctx *config.RunContext)) {
	t.Helper()

	runCtx, err := config.NewRunContextFromEnv(context.Background())
	assert.NoError(t, err)

	for _, ctxOption := range ctxOptions {
		ctxOption(runCtx)
	}

	level := "warn"
	if options.LogLevel != nil {
		level = *options.LogLevel
	}

	logBuf := testutil.ConfigureTestToCaptureLogs(t, runCtx, level)

	if options != nil && options.Currency != "" {
		runCtx.Config.Currency = options.Currency
	}

	require.NoError(t, err)

	// Load the arm projects
	armProjectData, err := os.ReadFile(filepath.Join("testdata", testName, testName+".json"))
	require.NoError(t, err)
	armProject := ARMProject{
		Files: []File{
			{
				Path:     "main.json",
				Contents: string(armProjectData),
			},
		},
	}

	// Load the usage data, if any.
	var usageData schema.UsageMap
	usageFilePath := filepath.Join("testdata", testName, testName+".usage.yml")
	if _, err := os.Stat(usageFilePath); err == nil || !os.IsNotExist(err) {
		// usage file exists, load the data
		usageFile, err := usage.LoadUsageFile(usageFilePath)
		require.NoError(t, err)
		usageData = usageFile.ToUsageDataMap()
	}

	projects := loadResources(t, armProject, runCtx, usageData)

	// Generate the output
	projects, err = RunCostCalculations(runCtx, projects)
	require.NoError(t, err)

	r, err := output.ToOutputFormat(runCtx.Config, projects)
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

	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

func loadResources(t *testing.T, armProject ARMProject, runCtx *config.RunContext, usageData schema.UsageMap) []*schema.Project {
	t.Helper()

	armdir := createARMProject(t, armProject)
	runCtx.Config.RootPath = armdir

	provider := arm.NewTemplateProvider(config.NewProjectContext(runCtx, &config.Project{
		Path: path.Join(armdir, "main.json"),
	}, nil), false, path.Join(armdir, "main.json"))

	projects, err := provider.LoadResources(usageData)
	require.NoError(t, err)

	for _, project := range projects {
		project.Name = strings.ReplaceAll(project.Name, armdir, t.Name())
		project.Name = strings.ReplaceAll(project.Name, "/arm", "")
		project.Name = strings.ReplaceAll(project.Name, "/main.json", "")
		project.BuildResources(schema.UsageMap{})
	}

	return projects
}

func RunCostCalculations(runCtx *config.RunContext, projects []*schema.Project) ([]*schema.Project, error) {
	pf := prices.NewPriceFetcher(runCtx)
	for _, project := range projects {
		err := pf.PopulatePrices(project)
		if err != nil {
			return projects, err
		}

		schema.CalculateCosts(project)
	}

	return projects, nil
}

func CreateARMProject(tmpDir string, armProject ARMProject) (string, error) {
	return writeToTmpDir(tmpDir, armProject)
}

func createARMProject(t *testing.T, armProject ARMProject) string {
	t.Helper()
	tmpDir := t.TempDir()

	armdir, err := CreateARMProject(tmpDir, armProject)
	require.NoError(t, err)

	return armdir
}

func writeToTmpDir(tmpDir string, armProject ARMProject) (string, error) {
	var err error

	for _, armFile := range armProject.Files {
		fullPath := filepath.Join(tmpDir, armFile.Path)
		dir := filepath.Dir(fullPath)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return tmpDir, err
			}
		}

		err = os.WriteFile(fullPath, []byte(armFile.Contents), 0600)
		if err != nil {
			return tmpDir, err
		}
	}

	return tmpDir, err
}
