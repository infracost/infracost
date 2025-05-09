package aws_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
)

func TestSFnStateMachineGoldenFile(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Don't run the terraform cli since it is complaining about the test examples:
	// Error: validating Step Functions State Machine definition:
	//        operation error SFN: ValidateStateMachineDefinition,
	//        https response error StatusCode: 400, RequestID: 4059cfcd-96ae-4921-afb0-25a552d4f601,
	//        api error UnrecognizedClientException: The security token included in the request is invalid.
	opts := tftest.DefaultGoldenFileOptions()
	opts.IgnoreCLI = true

	tftest.GoldenFileHCLResourceTestsWithOpts(t, "sfn_state_machine_test", opts)
}
