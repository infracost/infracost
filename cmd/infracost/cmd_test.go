package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	main "github.com/infracost/infracost/cmd/infracost"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/testutil"
)

var (
	timestampRegex   = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})(T| )(\d{2}):(\d{2}):(\d{2}(?:\.\d*)?)(([\+-](\d{2}):(\d{2})|Z| [A-Z]+)?)`)
	outputPathRegex  = regexp.MustCompile(`Output saved to .*`)
	urlRegex         = regexp.MustCompile(`https://dashboard.infracost.io/share/.*`)
	projectPathRegex = regexp.MustCompile(`(Project: .*) \(.*/infracost/infracost/examples/.*\)`)
)

type GoldenFileOptions = struct {
	Currency    string
	CaptureLogs bool
	IsJSON      bool
}

func DefaultOptions() *GoldenFileOptions {
	return &GoldenFileOptions{
		Currency:    "USD",
		CaptureLogs: false,
		IsJSON:      false,
	}
}

func GoldenFileCommandTest(t *testing.T, testName string, args []string, testOptions *GoldenFileOptions, configOptions ...func(*config.Config)) {
	if testOptions == nil {
		testOptions = DefaultOptions()
	}

	// Fix the VCS repo URL so the golden files don't fail on forks
	os.Setenv("INFRACOST_VCS_REPOSITORY_URL", "https://github.com/infracost/infracost")
	os.Setenv("INFRACOST_VCS_PULL_REQUEST_URL", "NOT_APPLICABLE")

	runCtx, err := config.NewRunContextFromEnv(context.Background())
	require.Nil(t, err)

	errBuf := bytes.NewBuffer([]byte{})
	outBuf := bytes.NewBuffer([]byte{})

	runCtx.Config.EventsDisabled = true
	runCtx.Config.Currency = testOptions.Currency
	runCtx.Config.NoColor = true

	for _, f := range configOptions {
		f(runCtx.Config)
	}

	rootCmd := main.NewRootCommand(runCtx)
	rootCmd.SetErr(errBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetArgs(args)

	var logBuf *bytes.Buffer
	if testOptions.CaptureLogs {
		logBuf = testutil.ConfigureTestToCaptureLogs(t, runCtx)
	} else {
		testutil.ConfigureTestToFailOnLogs(t, runCtx)
	}

	var cmdErr error
	var actual []byte

	cmdErr = rootCmd.Execute()

	if testOptions.IsJSON {
		prettyBuf := bytes.NewBuffer([]byte{})
		err = json.Indent(prettyBuf, outBuf.Bytes(), "", "  ")
		require.Nil(t, err)
		actual = prettyBuf.Bytes()
	} else {
		actual = outBuf.Bytes()
	}

	var errBytes []byte

	if errBuf != nil && errBuf.Len() > 0 {
		errBytes = append(errBytes, errBuf.Bytes()...)
	}

	if cmdErr != nil {
		errBytes = append(errBytes, []byte("Error: ")...)
		errBytes = append(errBytes, cmdErr.Error()...)
	}

	if len(errBytes) > 0 {
		actual = append(actual, "\nErr:\n"...)
		actual = append(actual, errBytes...)
	}

	if logBuf != nil && logBuf.Len() > 0 {
		actual = append(actual, "\nLogs:\n"...)
		actual = append(actual, logBuf.Bytes()...)
	}

	actual = stripDynamicValues(actual)

	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}

// stripDynamicValues strips out any values that change between test runs from the output,
// including timestamps and temp file paths
func stripDynamicValues(actual []byte) []byte {
	actual = timestampRegex.ReplaceAll(actual, []byte("REPLACED_TIME"))
	actual = outputPathRegex.ReplaceAll(actual, []byte("Output saved to REPLACED_OUTPUT_PATH"))
	actual = urlRegex.ReplaceAll(actual, []byte("https://dashboard.infracost.io/share/REPLACED_SHARE_CODE"))
	actual = projectPathRegex.ReplaceAll(actual, []byte("$1 REPLACED_PROJECT_PATH"))

	return actual
}
