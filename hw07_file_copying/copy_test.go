package main

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

const testsFileContent = "0123456789"

type RealTest func(t *testing.T, tempFilePath string)

func WithTempFileContent(t *testing.T, test RealTest) {
	tempFile, err := ioutil.TempFile("./testdata", "testsInputFile")
	require.NoErrorf(t, err, "unable to create temp file")
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			require.NoErrorf(t, err, "unable to delete temp file %s", tempFile.Name())
		}
	}()

	_, err = tempFile.WriteString(testsFileContent)
	tempFile.Close()

	require.NoError(t, err)
	test(t, tempFile.Name())
}

func TestCopyWholeFile(t *testing.T) {
	WithTempFileContent(t, func(t *testing.T, tempFilePath string) {
		copyOutputPath := path.Join(".", t.Name())

		err := Copy(tempFilePath, copyOutputPath, 0, int64(len(testsFileContent)))
		require.NoError(t, err)
		defer func() {
			err := os.Remove(copyOutputPath)
			require.NoError(t, err)
		}()

		content, err := ioutil.ReadFile(copyOutputPath)
		require.NoError(t, err)
		require.Equal(t, testsFileContent, string(content))
	})
}

func TestCopyFirst50PercentOfFile(t *testing.T) {
	WithTempFileContent(t, func(t *testing.T, tempFilePath string) {
		copyOutputPath := path.Join(".", t.Name())
		err := Copy(tempFilePath, copyOutputPath, 0, int64(len(testsFileContent))/2)
		require.NoError(t, err)
		defer func() {
			err := os.Remove(copyOutputPath)
			require.NoError(t, err)
		}()

		content, err := ioutil.ReadFile(copyOutputPath)
		require.NoError(t, err)
		require.Equal(t, testsFileContent[:int64(len(testsFileContent))/2], string(content))
	})
}

func TestCopyLast50PercentOfFile(t *testing.T) {
	WithTempFileContent(t, func(t *testing.T, tempFilePath string) {
		copyOutputPath := path.Join(".", t.Name())
		err := Copy(tempFilePath, copyOutputPath, int64(len(testsFileContent))/2, 0)
		require.NoError(t, err)
		defer func() {
			err := os.Remove(copyOutputPath)
			require.NoError(t, err)
		}()

		content, err := ioutil.ReadFile(copyOutputPath)
		require.NoError(t, err)
		require.Equal(t, testsFileContent[int64(len(testsFileContent))/2:], string(content))
	})
}

func TestFromPathIsWrong(t *testing.T) {
	err := Copy("blabla", "./testdata/output.txt", 0, 0)
	require.EqualError(t, err, "error during opening blabla file: open blabla: no such file or directory")
}

func TestOffsetBiggerThanFile(t *testing.T) {
	stat, err := os.Stat("./testdata/input.txt")
	require.NoError(t, err)
	err = Copy("./testdata/input.txt", "./testdata/output.txt", stat.Size()*2, 0)
	require.EqualError(t, err, ErrOffsetExceedsFileSize.Error())
}
