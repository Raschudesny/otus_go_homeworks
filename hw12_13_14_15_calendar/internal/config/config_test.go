package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	tempDirectoryPath = "./"
	tempDirName       = "tempdir"
	tempFileName      = "tempconfig"
	testConfigContent = `
logger:
  level: info
  file: some-log-output
api:
  http:
    port: 1234
    connectionTimeout: 10
  grpc:
    port: 56789
    connectionTimeout: 10
storage:
  inMemoryStorage: true
  db:
    host: some-awesome-postgres-url
    port: 12345
    username: zloygopnik123
    password: qwerty
    db: calendar`
)

func TestConfigReading(t *testing.T) {
	// create temp directory
	tempDir, err := ioutil.TempDir(tempDirectoryPath, tempDirName)
	require.NoErrorf(t, err, "unable to create temp directory")
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			require.NoErrorf(t, err, "unable to delete temp dir %s", tempDir)
		}
	}()
	// crete temp file in temp directory
	tempFile, err := ioutil.TempFile(tempDir, tempFileName)
	require.NoErrorf(t, err, "unable to create temp file")
	defer func() {
		if err := os.Remove(tempFile.Name()); err != nil {
			require.NoErrorf(t, err, "unable to delete temp file %s", tempFile.Name())
		}
	}()

	if _, err = tempFile.WriteString(testConfigContent); err != nil {
		require.NoError(t, err, "unable to write content to file %s", tempFile.Name())
	}

	if err := tempFile.Close(); err != nil {
		require.NoErrorf(t, err, "unable to close temp file %s", tempFile.Name())
	}

	config, err := NewConfig(tempFile.Name())
	require.NoError(t, err)

	require.Equal(t, "info", config.Logger.Level)
	require.Equal(t, "some-log-output", config.Logger.File)
	require.Equal(t, 1234, config.API.HTTP.Port)
	require.Equal(t, 10, config.API.HTTP.ConnectionTimeout)
	require.Equal(t, 56789, config.API.GRPC.Port)
	require.Equal(t, 10, config.API.GRPC.ConnectionTimeout)
	require.True(t, config.Storage.UseMemoryStorage)
	require.Equal(t, 12345, config.Storage.DB.Port)
	require.Equal(t, "some-awesome-postgres-url", config.Storage.DB.Host)
	require.Equal(t, "zloygopnik123", config.Storage.DB.Username)
	require.Equal(t, "qwerty", config.Storage.DB.Password)
	require.Equal(t, "calendar", config.Storage.DB.DB)
}
