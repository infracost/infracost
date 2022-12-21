package test

import (
	"testing"

	"github.com/infracost/infracost/costtest"
)

func TestAWSInstanceWebApp(t *testing.T) {
	costtest.Run(t, "aws_instance.web_app", func(t2 *costtest.T, r *costtest.Resource) {
		t2.InDelta(750, r.MonthlyCost, 50)

		t2.InDelta(180, r.SubResources.Get("ebs_block_device[0]").MonthlyCost, 10)
	})
}

func TestAWSInstanceModule(t *testing.T) {
	module := costtest.Group(t, "module.app")

	module.Run(t, "aws_instance.module_app", func(t2 *costtest.T, r *costtest.Resource) {
		t2.InDelta(1300, r.MonthlyCost, 50)
	})
}
