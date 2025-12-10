package main_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestCompletionHelpFlag(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"completion", "--help"}, nil)
}

func TestCompletionShellBash(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"completion", "--shell", "bash"}, nil)
}

func TestCompletionShellZsh(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"completion", "--shell", "zsh"}, nil)
}

func TestCompletionShellFish(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"completion", "--shell", "fish"}, nil)
}

func TestCompletionShellPowershell(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"completion", "--shell", "powershell"}, nil)
}
