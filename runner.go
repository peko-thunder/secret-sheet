package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

// runCommand executes command with args, injecting envVars on top of the
// current process environment. stdio is inherited so output goes directly
// to the terminal. The function forwards the child's exit code.
func runCommand(command string, args []string, envVars map[string]string) error {
	// On Windows npm/python live as .cmd / .exe wrappers; run through cmd.exe
	// so PATH resolution works the same way as in a normal shell.
	if runtime.GOOS == "windows" {
		args = append([]string{"/C", command}, args...)
		command = "cmd"
	}

	cmd := exec.Command(command, args...)

	// Build environment: inherit everything, then layer our vars on top.
	// Using os.Environ() + appended entries means our values override
	// any identically-named variable that was already set.
	env := os.Environ()
	for k, v := range envVars {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		if isNotFound(err) {
			return fmt.Errorf("command not found: %s", command)
		}
		return err
	}

	// Forward interrupt / termination signals to the child process so that
	// Ctrl-C, SIGTERM etc. propagate correctly.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			if cmd.Process != nil {
				cmd.Process.Signal(sig) //nolint:errcheck
			}
		}
	}()

	waitErr := cmd.Wait()
	signal.Stop(sigCh)
	close(sigCh)

	if exitErr, ok := waitErr.(*exec.ExitError); ok {
		os.Exit(exitErr.ExitCode())
	}
	return waitErr
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*exec.Error)
	return ok
}
