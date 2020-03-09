package pkg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"io/ioutil"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertTextFileContentEqual asserts that the expected file contents are the same as the actual file contents
// showing a nice diff if they differ
func AssertTextFileContentEqual(t *testing.T, expectedFile, actualFile, testName string) {
	if assert.FileExists(t, expectedFile, testName) && assert.FileExists(t, actualFile, testName) {
		expectedData, err := ioutil.ReadFile(expectedFile)
		require.NoError(t, err, "failed to load file %s", expectedFile)

		actualData, err := ioutil.ReadFile(actualFile)
		require.NoError(t, err, "failed to load file %s", actualFile)

		diff := cmp.Diff(string(actualData), string(expectedData))
		if diff != "" {
			t.Logf("generated: %s does not match expected: %s", actualFile, expectedFile)
			t.Logf("%s\n", diff)
			assert.Fail(t, "file %s is not the same as file %s", actualFile, expectedFile)
		}
	}
}

// AssertHelmTemplate asserts that we can generate resources for the given chart and folder of values files
func AssertHelmTemplate(t *testing.T, chart string, outDir, valuesDir string) (string, []string, error) {
	fileNames := []string{}
	releaseName := "myrel"
	ns := "jx"

	helm2, err := checkIfHelm2(t)
	if err != nil {
		return "", nil, err
	}
	// lets check if we have a requirements file
	requirementsFile := filepath.Join(chart, "requirements.yaml")
	exists, err := FileExists(requirementsFile)
	if err == nil && exists {
		// lets fetch dependencies
		t.Logf("building helm dependencies/n")
		args := []string{"dependency", "build", chart}
		cmd := exec.Command("helm", args...)
		data, err := cmd.CombinedOutput()
		if data != nil {
			t.Logf("result: %s\n", string(data))
		}
		require.NoError(t, err, "failed to build dependencies: helm %s", strings.Join(args, " "))
	}

	args := []string{"template", releaseName, chart, "--output-dir", outDir, "--namespace", ns}
	if helm2 {
		t.Logf("using helm 2.x binary/n")
		args = []string{"template", "--name", releaseName, chart, "--output-dir", outDir}
	}

	files, err := ioutil.ReadDir(valuesDir)
	require.NoError(t, err, "could not read dir %s", valuesDir)
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".yaml") {
			args = append(args, "--values", filepath.Join(valuesDir, name))
		}
	}
	err = os.MkdirAll(outDir, DefaultDirWritePermissions)
	require.NoError(t, err, "failed to make output dir %s", outDir)

	commandLine := "helm " + strings.Join(args, " ")
	t.Logf("invoking: %s\n", commandLine)

	cmd := exec.Command("helm", args...)
	data, err := cmd.CombinedOutput()
	if data != nil {
		t.Logf("result: %s\n", string(data))
	}
	require.NoError(t, err, "failed to run: %s", commandLine)

	// now lets assert that the files are valid
	templatesDir, err := filepath.Glob(filepath.Join(outDir, "*", "templates"))
	require.NoError(t, err, "could not find templates dir in dir %s", outDir)
	require.NotEmpty(t, templatesDir, "no */templates dir in dir %s", outDir)
	templateDir := templatesDir[0]

	resultDir := filepath.Join(outDir, "results")

	files, err = ioutil.ReadDir(templateDir)
	for _, f := range files {
		name := f.Name()
		if strings.HasSuffix(name, ".yaml") {
			path := filepath.Join(templateDir, name)

			names, err := splitObjectsInFiles(t, path, resultDir)
			require.NoError(t, err, "failed to split the yaml file into resulting files for %s", path)
			fileNames = append(fileNames, names...)
		}
	}
	return resultDir, fileNames, err
}
