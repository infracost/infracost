package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"

	"github.com/shopspring/decimal"
)

func TestCodebuildProject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_codebuild_project" "my_project" {
			name           = "test-project-cache"
  		description    = "test_codebuild_project_cache"
			
			service_role = ""

			artifacts {
				type = "NO_ARTIFACTS"
			}

			environment {
				compute_type                = "BUILD_GENERAL1_MEDIUM"
				image                       = "aws/codebuild/standard:1.0"
				type                        = "LINUX_CONTAINER"
				image_pull_credentials_type = "CODEBUILD"
			}

			source {
				type            = "GITHUB"
				location        = ""
				git_clone_depth = 1
			}
		}`

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_codebuild_project.my_project",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Linux (general1.medium)",
					PriceHash:        "a26b218d7a04b4de7dc49fc899fcbf7f-a62d9273fef0987b8d1b9a67a508acdc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.Zero),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, schema.NewEmptyUsageMap(), resourceChecks)
}

func TestCodebuildProject_usage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
		resource "aws_codebuild_project" "my_small_project" {
			name           = "test-project-cache"
			description    = "test_codebuild_project_cache"
			
			service_role = ""

			artifacts {
				type = "NO_ARTIFACTS"
			}

			environment {
				compute_type                = "BUILD_GENERAL1_SMALL"
				image                       = "aws/codebuild/standard:1.0"
				type                        = "LINUX_CONTAINER"
				image_pull_credentials_type = "CODEBUILD"
			}

			source {
				type            = "GITHUB"
				location        = ""
				git_clone_depth = 1
			}
		}

		resource "aws_codebuild_project" "my_medium_project" {
			name           = "test-project-cache"
  		description    = "test_codebuild_project_cache"
			
			service_role = ""

			artifacts {
				type = "NO_ARTIFACTS"
			}

			environment {
				compute_type                = "BUILD_GENERAL1_MEDIUM"
				image                       = "aws/codebuild/standard:1.0"
				type                        = "LINUX_CONTAINER"
				image_pull_credentials_type = "CODEBUILD"
			}

			source {
				type            = "GITHUB"
				location        = ""
				git_clone_depth = 1
			}
		}

		resource "aws_codebuild_project" "my_large_linux_project" {
			name           = "test-project-cache"
  		description    = "test_codebuild_project_cache"
			
			service_role = ""

			artifacts {
				type = "NO_ARTIFACTS"
			}

			environment {
				compute_type                = "BUILD_GENERAL1_LARGE"
				image                       = "aws/codebuild/standard:1.0"
				type                        = "LINUX_CONTAINER"
				image_pull_credentials_type = "CODEBUILD"
			}

			source {
				type            = "GITHUB"
				location        = ""
				git_clone_depth = 1
			}
		}
		
		resource "aws_codebuild_project" "my_large_windows_project" {
			name           = "test-project-cache"
  		description    = "test_codebuild_project_cache"
			
			service_role = ""

			artifacts {
				type = "NO_ARTIFACTS"
			}

			environment {
				compute_type                = "BUILD_GENERAL1_LARGE"
				image                       = "aws/codebuild/standard:1.0"
				type                        = "WINDOWS_SERVER_2019_CONTAINER"
				image_pull_credentials_type = "CODEBUILD"
			}

			source {
				type            = "GITHUB"
				location        = ""
				git_clone_depth = 1
			}
		}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"aws_codebuild_project.my_small_project": map[string]interface{}{
			"monthly_build_mins": 1000,
		},
		"aws_codebuild_project.my_medium_project": map[string]interface{}{
			"monthly_build_mins": 10000,
		},
		"aws_codebuild_project.my_large_linux_project": map[string]interface{}{
			"monthly_build_mins": 100000,
		},
		"aws_codebuild_project.my_large_windows_project": map[string]interface{}{
			"monthly_build_mins": 100000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "aws_codebuild_project.my_small_project",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Linux (general1.small)",
					PriceHash:        "78647b140df3f8c5350ab75213cac828-a62d9273fef0987b8d1b9a67a508acdc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(1000)),
				},
			},
		},
		{
			Name: "aws_codebuild_project.my_medium_project",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Linux (general1.medium)",
					PriceHash:        "a26b218d7a04b4de7dc49fc899fcbf7f-a62d9273fef0987b8d1b9a67a508acdc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(10000)),
				},
			},
		},
		{
			Name: "aws_codebuild_project.my_large_linux_project",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Linux (general1.large)",
					PriceHash:        "05233b2fb94a8929a2bc26c8a4000b1c-a62d9273fef0987b8d1b9a67a508acdc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100000)),
				},
			},
		},
		{
			Name: "aws_codebuild_project.my_large_windows_project",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:             "Windows (general1.large)",
					PriceHash:        "a5080472369b82f5143a4c9a5b1381ee-a62d9273fef0987b8d1b9a67a508acdc",
					MonthlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromFloat(100000)),
				},
			},
		},
	}

	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
