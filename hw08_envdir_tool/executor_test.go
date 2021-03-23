package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmdWithOnlySystem(t *testing.T) {
	var mockedStdout bytes.Buffer
	resCode := RunCmd([]string{"env"}, nil, CmdIOStreams{in: os.Stdin, out: &mockedStdout, err: os.Stderr})
	require.Equal(t, 0, resCode)
	for _, envEntry := range os.Environ() {
		require.Contains(t, mockedStdout.String(), envEntry)
	}
}

func TestRunCmdEraseAll(t *testing.T) {
	systemEnvs := os.Environ()
	erasingEnvs := make(Environment, len(systemEnvs))
	for _, envEntry := range systemEnvs {
		envParts := strings.Split(envEntry, "=")
		if len(envParts) >= 1 {
			envKey := envParts[0]
			erasingEnvs[envKey] = EnvValue{"", true}
		}
	}

	var mockedStdout bytes.Buffer
	resCode := RunCmd([]string{"env"}, erasingEnvs, CmdIOStreams{in: os.Stdin, out: &mockedStdout, err: os.Stderr})
	require.Equal(t, 0, len(mockedStdout.String()))
	require.Equal(t, 0, resCode)
}

func TestRunCmdWithOneEnv(t *testing.T) {
	var mockedStdout bytes.Buffer
	envs := make(Environment, 1)
	envs["custom_env"] = EnvValue{"custom_env_value", false}
	resCode := RunCmd([]string{"env"}, envs, CmdIOStreams{in: os.Stdin, out: &mockedStdout, err: os.Stderr})
	require.Equal(t, 0, resCode)
	require.Contains(t, mockedStdout.String(), "custom_env=custom_env_value")
}
