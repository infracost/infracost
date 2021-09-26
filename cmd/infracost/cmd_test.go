package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	main "github.com/infracost/infracost/cmd/infracost"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/stretchr/testify/require"
)

var timestampRegex = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})(T| )(\d{2}):(\d{2}):(\d{2}(?:\.\d*)?)(([\+-](\d{2}):(\d{2})|Z| [A-Z]{3})?)`)

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

func GoldenFileCommandTest(t *testing.T, testName string, args []string, options *GoldenFileOptions) {
	if options == nil {
		options = DefaultOptions()
	}

	// Fix the VCS repo URL so the golden files don't fail on forks
	os.Setenv("INFRACOST_VCS_REPOSITORY_URL", "https://github.com/infracost/infracost.git")

	runCtx, err := config.NewRunContextFromEnv(context.Background())
	require.Nil(t, err)

	errBuf := bytes.NewBuffer([]byte{})
	outBuf := bytes.NewBuffer([]byte{})

	runCtx.Config.EventsDisabled = true
	runCtx.Config.Currency = options.Currency
	runCtx.Config.NoColor = true

	rootCmd := main.NewRootCommand(runCtx)
	rootCmd.SetErr(errBuf)
	rootCmd.SetOut(outBuf)
	rootCmd.SetArgs(args)

	var logBuf *bytes.Buffer
	if options.CaptureLogs {
		logBuf = testutil.ConfigureTestToCaptureLogs(t, runCtx)
	} else {
		testutil.ConfigureTestToFailOnLogs(t, runCtx)
	}

	var cmdErr error
	var actual []byte

	cmdErr = rootCmd.Execute()

	if options.IsJSON {
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

	// strip out any timestamps
	actual = timestampRegex.ReplaceAll(actual, []byte("REPLACED_TIME"))

	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}
