package main

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testsDirectory = "./testdata"

func createTempFileWithContent(t *testing.T, content string) string {
	file, err := ioutil.TempFile(testsDirectory, "tempfile")
	require.NoErrorf(t, err, "unable to create temp file")
	_, err = file.WriteString(content)
	file.Close()
	require.NoError(t, err)
	return file.Name()
}

func removeTempTestFile(t *testing.T, fileName string) {
	if err := os.Remove(fileName); err != nil {
		require.NoErrorf(t, err, "unable to delete temp file %s", fileName)
	}
}

func TestReadDirEnvs(t *testing.T) {
	envs, err := ReadDirEnvs("./testdata/env")
	require.Len(t, envs, 5)
	require.NoError(t, err)
}

func TestReadNotExistDir(t *testing.T) {
	dir, err := ReadDirEnvs("./some_not_valid_dir_name")
	// wow, c'mon golang why should i cast nil to Environment here just to force this check pass
	require.Equal(t, Environment(nil), dir)
	var pathErr *os.PathError
	require.True(t, errors.As(err, &pathErr))
}

func TestReadEmptyDir(t *testing.T) {
	tempDir, err := ioutil.TempDir("./testdata", "tempdir")
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			require.NoErrorf(t, err, "unable to delete temp dir %s", tempDir)
		}
	}()
	require.NoError(t, err)
	dirEnvs, err := ReadDirEnvs(tempDir)
	require.NoError(t, err)
	// wow, c'mon golang why should i cast nil to Environment here just to force this check pass
	require.Equal(t, Environment(nil), dirEnvs)
}

func TestReadFirstLineFromFile(t *testing.T) {
	for _, testData := range [...]struct {
		filename string
		expected string
		name     string
	}{
		{
			createTempFileWithContent(t, ""),
			"",
			"empty file",
		},
		{
			createTempFileWithContent(t, "12345"),
			"12345",
			"one line file",
		},
		{
			createTempFileWithContent(t, "123\n123"),
			"123\n",
			"two line file",
		},
	} {
		t.Run(testData.name, func(t *testing.T) {
			defer removeTempTestFile(t, testData.filename)
			line, err := ReadFirstLineFromFile(testData.filename)
			require.NoError(t, err)
			require.Equal(t, testData.expected, line)
		})
	}
}
