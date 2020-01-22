package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCase represents an individual test case
type TestCase struct {
	Chart                string
	Name                 string
	ValuesDir            string
	ExpectedDir          string
	OutDir               string
	ExpectFail           bool
	FailOnExtraResources bool
	t                    *testing.T
}

// Run runs the unit test
func (c *TestCase) Run() {
	t := c.t
	testOutDir := c.OutDir
	resultsDir, _, err := AssertHelmTemplate(t, c.Chart, testOutDir, c.ValuesDir)

	require.NoError(t, err, "failed to generate helm templates")
	require.NotEmpty(t, resultsDir, "no resultsDir returned")

	if GenerateExpectedFiles.Value() {
		c.RegenerateExpectedFiles(t, resultsDir)
		return
	}
	c.AssertYamlExpected(t, resultsDir)
}

// AssertYamlExpected asserts that the expectedDir of generated YAML is contained in the actualDir
func (c *TestCase) AssertYamlExpected(t *testing.T, actualDir string) {
	count := 0
	expectedFiles := map[string]bool{}
	expectedDir := c.ExpectedDir
	err := filepath.Walk(expectedDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".yaml" {
				t.Logf("testing expected file %s", path)

				relPath, err := filepath.Rel(expectedDir, path)
				if err != nil {
					return errors.Wrapf(err, "failed to get base path for %s", path)
				}
				expectedFiles[relPath] = true
				actualFile := filepath.Join(actualDir, relPath)
				AssertTextFileContentEqual(t, path, actualFile, c.Name)
				count++
			}
			return nil
		})
	require.NoError(t, err, "failed to verify expected files")
	t.Logf("verified %d files match the expected YAML files in %s\n", count, expectedDir)

	if c.FailOnExtraResources {
		extraResources := []string{}
		// lets verify that there were no extra resources created
		err = filepath.Walk(actualDir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && filepath.Ext(path) == ".yaml" {
					relPath, err := filepath.Rel(actualDir, path)
					if err != nil {
						return errors.Wrapf(err, "failed to get base path for %s", path)
					}
					if !expectedFiles[relPath] {
						extraResources = append(extraResources, relPath)
					}
				}
				return nil
			})

		if c.ExpectFail {
			t.Logf("as expected test case %s failed due to exta resources being generated that were not in the expected directory: %v", c.Name, extraResources)
			if err != nil {
				t.Logf("test case %s got expected error %s", c.Name, err.Error())
			}
		} else {
			assert.Empty(t, extraResources, "these resources were generated but are not in the expected resources directory %s", expectedDir)
			assert.NoError(t, err, "failed to verify no extra files were created")
		}
	}
}

// RegenerateExpectedFiles regenerates the expected files
func (c *TestCase) RegenerateExpectedFiles(t *testing.T, actualDir string) {
	expectedDir := c.ExpectedDir
	t.Logf("regenerating the expected files for test %s into dir %s", c.Name, expectedDir)
	count := 0
	err := filepath.Walk(actualDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".yaml" {
				relPath, err := filepath.Rel(actualDir, path)
				if err != nil {
					return errors.Wrapf(err, "failed to get base path for %s", path)
				}
				expectedFile := filepath.Join(expectedDir, relPath)
				err = CopyFile(path, expectedFile)
				if err != nil {
					return errors.Wrapf(err, "failed to copy result to expected file %s", expectedFile)
				}
				count++
			}
			return nil
		})
	require.NoError(t, err, "failed to regenerate expected files")
	t.Logf("regenerated %d expected files mto %s\n", count, expectedDir)
}
