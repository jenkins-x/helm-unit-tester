package pkg_test

import (
	"path/filepath"
	"testing"

	"github.com/jenkins-x/helm-unit-tester/pkg"
)

func TestChartsWithDifferentValues(t *testing.T) {
	chart, valid := pkg.AssertChartPathExists(t, filepath.Join("test_data", "testchart"))
	if valid {
		pkg.RunTests(t, chart, filepath.Join("test_data", "tests"))
	}
}
