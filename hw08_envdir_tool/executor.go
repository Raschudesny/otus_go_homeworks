package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// this type in hw08 is mostly needed to simplify RunCmd tests.
type CmdIOStreams struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

// RunCmd runs a command(cmd[0]) + arguments (cmd[1:]) with environment variables from env.
func RunCmd(cmd []string, dirEnvs Environment, ioStreams CmdIOStreams) (returnCode int) {
	command := exec.Command(cmd[0], cmd[1:]...) // #nosec G204
	command.Stdout = ioStreams.out
	command.Stdin = ioStreams.in
	command.Stderr = ioStreams.err

	systemEnvs := os.Environ()
	resultEnvs := make([]string, 0, len(systemEnvs)+len(dirEnvs))

	for _, envEntry := range systemEnvs {
		envParts := strings.Split(envEntry, "=")
		if len(envParts) >= 2 {
			osEnvKey := envParts[0]
			osEnvValue := envParts[1]
			if dirEnvVal, ok := dirEnvs[osEnvKey]; ok && dirEnvVal.NeedRemove {
				continue
			}
			resultEnvs = append(resultEnvs, strings.Join([]string{osEnvKey, osEnvValue}, "="))
		}
	}

	for k, v := range dirEnvs {
		if !v.NeedRemove {
			resultEnvs = append(resultEnvs, strings.Join([]string{k, v.Value}, "="))
		}
	}

	command.Env = resultEnvs
	err := command.Start()
	if err != nil {
		log.Fatalln(fmt.Errorf("unable to run command %s error: %w", cmd[0], err))
	}
	if err = command.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		log.Println(fmt.Errorf("error during command running: %w", err))
		return -1
	}
	return 0
}
