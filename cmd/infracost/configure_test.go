package main_test

import (
	"github.com/infracost/infracost/internal/testutil"
	"testing"
)

// We don't have a great way to test this, so just test the help.

func TestConfigureNoArgs(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"configure"}, nil)
}

func TestConfigureHelpFlag(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"configure", "--help"}, nil)
}
