package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/infracost/infracost/cmd/infracost"
	"github.com/infracost/infracost/internal/config"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"regexp"
	"testing"
)

var timestampRegex = regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})(T| )(\d{2}):(\d{2}):(\d{2}(?:\.\d*)?)(([\+-](\d{2}):(\d{2})|Z| [A-Z]{3})?)`)
var vcsRepoURLRegex = regexp.MustCompile(`"vcsRepoUrl": "[^"]*"`)

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

	err = rootCmd.Execute()
	require.Nil(t, err)

	var actual []byte

	if options.IsJSON {
		prettyBuf := bytes.NewBuffer([]byte{})
		err = json.Indent(prettyBuf, outBuf.Bytes(), "", "  ")
		require.Nil(t, err)
		actual = prettyBuf.Bytes()
	} else {
		actual = outBuf.Bytes()
	}

	if errBuf != nil && errBuf.Len() > 0 {
		actual = append(actual, "\nErr:\n"...)
		actual = append(actual, errBuf.Bytes()...)
	}

	if logBuf != nil && logBuf.Len() > 0 {
		actual = append(actual, "\nLogs:\n"...)
		actual = append(actual, logBuf.Bytes()...)
	}

	// strip out any timestamps
	actual = timestampRegex.ReplaceAll(actual, []byte("REPLACED_TIME"))
	actual = vcsRepoURLRegex.ReplaceAll(actual, []byte(`"vcsRepoUrl": "REPLACED"`))

	goldenFilePath := filepath.Join("testdata", testName, testName+".golden")
	testutil.AssertGoldenFile(t, goldenFilePath, actual)
}
