package pkg

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

// UnitTester a unit tester of charts
type UnitTester struct {
	// OutDir the directory charts are generated to
	OutDir string

	t *testing.T
}

// NewUnitTester creates a new unit tester of charts
func NewUnitTester(t *testing.T) (*UnitTester, error) {
	outDir, err := ioutil.TempDir("", "helm-test-")
	t.Logf("writing generated helm templates to %s", outDir)

	if err != nil {
		return nil, errors.Wrap(err, "could not create temp dir")
	}
	return &UnitTester{OutDir: outDir, t: t}, nil
}

// RunTests creates a tester for the given chart and test directories and runs the tests
func RunTests(t *testing.T, chart string, testDir string) (*UnitTester, []*TestCase) {
	tester, err := NewUnitTester(t)
	require.NoError(t, err, "failed to create chart UnitTester")

	tests, err := tester.LoadTests(chart, testDir)
	require.NoError(t, err, "failed to load tests")

	tester.RunTests(tests)
	return tester, tests
}

// AssertChartPath resolves the given path to an absolute file so that it is interpreted as a local file
// rather than a chart name and asserts the dir exists
func AssertChartPathExists(t *testing.T, path string) (string, bool) {
	chart, err := filepath.Abs(path)
	worked := assert.NoError(t, err, "failed to find absolute chart")
	if worked {
		worked = assert.DirExists(t, chart, "test chart does not exist")
	}
	return chart, worked
}

// RunTests finds all the tests inside the given testDir and runs the test for each one
func (u *UnitTester) LoadTests(chart string, testDir string) ([]*TestCase, error) {
	t := u.t
	tests := []*TestCase{}
	files, err := ioutil.ReadDir(testDir)
	require.NoError(t, err, "could not read dir %s", testDir)
	for _, f := range files {
		if f.IsDir() {
			name := f.Name()
			valuesDir := filepath.Join(testDir, name, "values")
			expectedDir := filepath.Join(testDir, name, "expected")

			testOutDir := filepath.Join(u.OutDir, name)

			testCase, err := u.loadTestCaseConfig(testDir, name)
			if err != nil {
				return nil, err
			}

			testCase.t = t
			testCase.Name = name
			testCase.Chart = chart
			testCase.ValuesDir = valuesDir
			testCase.ExpectedDir = expectedDir
			testCase.OutDir = testOutDir
			tests = append(tests, testCase)
		}
	}
	return tests, nil
}

func (u *UnitTester) loadTestCaseConfig(testDir string, name string) (*TestCase, error) {
	testCase := &TestCase{}
	configFile := filepath.Join(testDir, name, "testcase.yml")
	exists, err := FileExists(configFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to check for config file %s", configFile)
	}
	if exists {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load config file %s", configFile)
		}
		err = yaml.Unmarshal(data, testCase)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal config file %s", configFile)
		}
	}
	return testCase, nil
}

// RunTests finds all the tests inside the given testDir and runs the test for each one
func (u *UnitTester) RunTests(tests []*TestCase) {
	for _, utest := range tests {
		utest.t = u.t
		utest.Run()
	}
}
