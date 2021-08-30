package executor

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

/* exec.Command() does return a stdout pipe. But it blocks on read if no data is available.
   We need to create a workaround with a go thread reading it constantly,
   and checking its resulting buffer from time to time.
*/

type InteractiveShell struct {
	execCmd   *exec.Cmd // nil means closed
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	stdoutBuf *bytes.Buffer
	stderrBuf *bytes.Buffer
}

func makeInteractiveShell() InteractiveShell {
	interactiveShell := InteractiveShell{
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	}
	return interactiveShell
}

func (i *InteractiveShell) AlreadyOpen() bool {
	if i.execCmd == nil {
		return false
	} else {
		return true
	}
}

func (i *InteractiveShell) close() error {
	var execCmd = i.execCmd

	i.execCmd = nil
	i.stdin = nil
	i.stdout = nil
	i.stderr = nil
	i.stdoutBuf = nil
	i.stderrBuf = nil

	if execCmd == nil {
		return nil
	}

	return execCmd.Process.Kill()
}

func (i *InteractiveShell) open(executable string, args []string) (string, string, error) {
	log.Debugf("Starting interactive shell: %s %v", executable, args)
	exeCommand := exec.Command(executable, args...)
	stdin, err := exeCommand.StdinPipe()
	if err != nil {
		return "", "", err
	}
	stdout, err := exeCommand.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderr, err := exeCommand.StderrPipe()
	if err != nil {
		return "", "", err
	}

	// Start it
	if err := exeCommand.Start(); err != nil {
		return "", "", err
	}

	// Read initial stdin
	// Its always two read's. If not, it will block forever
	line1 := make([]byte, 100)
	n, err := stdout.Read(line1)
	if err != nil {
		return "", "", err
	}
	stdoutCut1 := line1[0:n]
	line2 := make([]byte, 100)
	n, err = stdout.Read(line2)
	if err != nil {
		return "", "", err
	}
	stdoutCut2 := line2[0:n]

	// Read initial stderr
	// Doesnt work, its usually empty so block forever
	/*line3 := make([]byte, 100)
	_, err = stderr.Read(line2)
	if err != nil {
		return "", "", err
	}*/

	// No errors till here, set it
	i.execCmd = exeCommand
	i.stdin = stdin
	i.stdout = stdout
	i.stderr = stderr
	i.stdoutBuf = bytes.NewBuffer(nil)
	i.stderrBuf = bytes.NewBuffer(nil)

	// read the stdout continuously in a separate goroutine and capture it in our vars
	// Read() will block if no data is available. Lifetime: app?
	go func() {
		for {
			part := make([]byte, 128)
			n, err := i.stdout.Read(part)
			if err != nil {
				break
			}
			i.stdoutBuf.Write(part[0:n])
		}
	}()
	go func() {
		for {
			part := make([]byte, 128)
			n, err := i.stderr.Read(part)
			if err != nil {
				break
			}
			i.stderrBuf.Write(part[0:n])
		}
	}()

	return string(stdoutCut1) + string(stdoutCut2), string(""), nil
}

func (i *InteractiveShell) issue(commandline string) (string, string, error) {
	if i.execCmd == nil || i.stdin == nil {
		return "", "", fmt.Errorf("Shell not open")
	}

	// Give command to packet
	// Do it every time, or we will block! (even when empty "")
	_, err := fmt.Fprintln(i.stdin, commandline)
	if err != nil {
		// process is most likely exited, handle it as such
		i.execCmd = nil
		i.stdin = nil
		i.stdout = nil
		i.stderr = nil
		i.stdoutBuf = bytes.NewBuffer(nil)
		i.stderrBuf = bytes.NewBuffer(nil)
		return "", "", fmt.Errorf("Shell down: %s", err.Error())
	}

	time.Sleep(100 * time.Millisecond) // Always give 100ms first
	/* We read until the output buffer size does not increase for a certain
	   amount of time (max 1s).
	   We cannot be sure if the process dumped all of its data, but thats  how it is.
	*/
	prevLen := 0
	n := 10
	for {
		n -= 1
		time.Sleep(100 * time.Millisecond)

		if n == 0 {
			// Max count reached (long lasting process with lots of output?)
			break
		}
		len := i.stdoutBuf.Len()
		if len == 0 {
			// No new data, wait for it
			continue
		}
		if len == prevLen {
			// Same amount of data after a sleep, lets take it
			break
		}
		prevLen = len
	}

	// Get data we aquired until now, and reset the buffers
	stdoutBytes := i.stdoutBuf.Bytes()
	i.stdoutBuf.Reset()
	stderrBytes := i.stderrBuf.Bytes()
	i.stderrBuf.Reset()
	stdoutStr := windowsToString(stdoutBytes)
	stderrStr := windowsToString(stderrBytes)

	return stdoutStr, stderrStr, nil
}
