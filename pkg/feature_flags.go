package pkg

import "github.com/jenkins-x/helm-unit-tester/pkg/flags"

var (
	// GenerateExpectedFiles if enabled lets copy the actual files generated over the expected files
	// e.g. if the underlying chart or helm has changed to reset the tests
	GenerateExpectedFiles = flags.NewBoolFlag(false, "HELM_UNIT_REGENERATE_EXPECTED")
)
