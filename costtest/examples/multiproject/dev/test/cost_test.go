package test

import (
	"testing"

	"github.com/infracost/infracost/costtest"
)

func TestMain(m *testing.M) {
	costtest.Init(costtest.InitOptions{Path: "../"})
	defer costtest.Cleanup()

	m.Run()
}
