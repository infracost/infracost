package main_test

import (
	"testing"

	"github.com/infracost/infracost/internal/testutil"
)

func TestComment(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"comment"}, nil)
}

func TestCommentHelp(t *testing.T) {
	GoldenFileCommandTest(t, testutil.CalcGoldenFileTestdataDirName(), []string{"comment", "--help"}, nil)
}
