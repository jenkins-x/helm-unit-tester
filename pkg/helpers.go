package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func splitObjectsInFiles(t *testing.T, inputFile string, resultDir string) ([]string, error) {
	fileNames := make([]string, 0)
	f, err := os.Open(inputFile)
	if err != nil {
		return fileNames, errors.Wrapf(err, "opening inputFile %q", inputFile)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if line == resourcesSeparator {
			// ensure that we actually have YAML in the buffer
			data := buf.Bytes()
			if isWhitespaceOrComments(data) {
				buf.Reset()
				continue
			}

			fileName, err := writeBufferToFile(t, data, resultDir)
			require.NoError(t, err, "failed to write buffer to file")
			if fileName != "" {
				fileNames = append(fileNames, fileName)
			}
			buf.Reset()
		} else {
			_, err := buf.WriteString(line)
			if err != nil {
				return fileNames, errors.Wrapf(err, "writing line from inputFile %q into a buffer", inputFile)
			}
			_, err = buf.WriteString("\n")
			if err != nil {
				return fileNames, errors.Wrapf(err, "writing a new line in the buffer")
			}
		}
	}
	if buf.Len() > 0 && !isWhitespaceOrComments(buf.Bytes()) {
		data := buf.Bytes()
		fileName, err := writeBufferToFile(t, data, resultDir)
		require.NoError(t, err, "failed to write buffer to file")
		if fileName != "" {
			fileNames = append(fileNames, fileName)
		}
	}
	return fileNames, nil
}

func writeBufferToFile(t *testing.T, data []byte, resultDir string) (string, error) {
	m := yaml.MapSlice{}
	err := yaml.Unmarshal(data, &m)
	require.NoError(t, err, "failed to unmarshal YAML: %s", string(data))

	if len(m) == 0 {
		return "", nil
	}

	name := getYamlValueString(&m, "metadata", "name")
	apiVersion := getYamlValueString(&m, "apiVersion")
	kind := getYamlValueString(&m, "kind")
	require.NotEmpty(t, name, "resource with missing name: %s", string(data))
	require.NotEmpty(t, kind, "resource with missing kind: %s", string(data))

	outDir := filepath.Join(resultDir, kind)
	if apiVersion != "" {
		outDir = filepath.Join(resultDir, apiVersion, kind)
	}
	err = os.MkdirAll(outDir, DefaultDirWritePermissions)
	require.NoError(t, err, "failed to make output dir %s", outDir)

	fileName := filepath.Join(outDir, name+".yaml")

	err = ioutil.WriteFile(fileName, data, DefaultFileWritePermissions)
	require.NoError(t, err, "creating file %q", fileName)
	return fileName, err
}

func getYamlValueString(mapSlice *yaml.MapSlice, keys ...string) string {
	value := getYamlValue(mapSlice, keys...)
	answer, ok := value.(string)
	if ok {

		return answer
	}
	return ""
}

func getYamlValue(mapSlice *yaml.MapSlice, keys ...string) interface{} {
	if mapSlice == nil {
		return nil
	}
	if mapSlice == nil {
		return fmt.Errorf("No map input!")
	}
	m := mapSlice
	lastIdx := len(keys) - 1
	for idx, k := range keys {
		last := idx >= lastIdx
		found := false
		for _, mi := range *m {
			if mi.Key == k {
				found = true
				if last {
					return mi.Value
				} else {
					value := mi.Value
					if value == nil {
						return nil
					} else {
						v, ok := value.(yaml.MapSlice)
						if ok {
							m = &v
						} else {
							v2, ok := value.(*yaml.MapSlice)
							if ok {
								m = v2
							} else {
								return nil
							}
						}
					}
				}
			}
		}
		if !found {
			return nil
		}
	}
	return nil
}

// isWhitespaceOrComments returns true if the data is empty, whitespace or comments only
func isWhitespaceOrComments(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t != "" && !strings.HasPrefix(t, "#") {
			return false
		}
	}
	return true
}

// FileExists checks if path exists and is a file
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, errors.Wrapf(err, "failed to check if file exists %s", path)
}

func checkIfHelm2(t *testing.T) (bool, error) {
	cmd := exec.Command("helm", "version", "-c", "--short")
	data, err := cmd.CombinedOutput()
	if data != nil {
		t.Logf("helm version: %s\n", string(data))
	}
	if err != nil {
		return false, err
	}

	fields := strings.Split(strings.TrimSpace(string(data)), " ")
	if len(fields) > 1 {
		v := strings.TrimPrefix(fields[1], "v")
		if strings.HasPrefix(v, "2.") {
			return true, nil
		}
	}
	return false, nil
}

// credit https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04
func CopyFile(src, dst string) (err error) {
	dstDir := filepath.Dir(dst)
	err = os.MkdirAll(dstDir, DefaultDirWritePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to create parent dir %s", dstDir)
	}
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}
	return
}
