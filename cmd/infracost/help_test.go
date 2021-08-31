package main_test

import (
	"github.com/infracost/infracost/internal/testutil"
	"testing"
)

func TestNoArgs(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{}, nil)
}

func TestHelpFlag(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"--help"}, nil)
}

func TestHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"help"}, nil)
}

func TestHelpHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"help", "--help"}, nil)
}
