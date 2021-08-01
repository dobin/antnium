package executor

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"time"
)

/* exec.Command() does return a stdout pipe. But it blocks on read if no data is available.
   We need to create a workaround with a go thread reading it constantly,
   and checking its resulting buffer from time to time.
*/

type InteractiveShell struct {
	execCmd   *exec.Cmd
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

func (interactiveShell *InteractiveShell) AlreadyOpen() bool {
	if interactiveShell.execCmd == nil {
		return false
	} else {
		return true
	}
}

func (interactiveShell *InteractiveShell) open() (string, string, error) {
	// Setup command
	exeCommand := exec.Command("cmd", "/a")
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
	interactiveShell.execCmd = exeCommand

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

	interactiveShell.stdin = stdin
	interactiveShell.stdout = stdout
	interactiveShell.stderr = stderr
	interactiveShell.stdoutBuf = bytes.NewBuffer(nil)
	interactiveShell.stderrBuf = bytes.NewBuffer(nil)

	// read the stdout continuously in a separate goroutine and capture it in our vars
	// Read() will block if no data is available.
	go func() {
		for {
			part := make([]byte, 128)
			n, err := interactiveShell.stdout.Read(part)
			if err != nil {
				break
			}
			interactiveShell.stdoutBuf.Write(part[0:n])
		}
	}()
	go func() {
		for {
			part := make([]byte, 128)
			n, err := interactiveShell.stderr.Read(part)
			if err != nil {
				break
			}
			interactiveShell.stderrBuf.Write(part[0:n])
		}
	}()

	return string(stdoutCut1) + string(stdoutCut2), string(""), nil
}

func (interactiveShell *InteractiveShell) issue(commandline string) (string, string, error) {
	if interactiveShell.stdin == nil {
		return "", "", fmt.Errorf("Shell not open")
	}

	// Give command to packet
	// Do it every time, or we will block! (even when empty "")
	fmt.Fprintln(interactiveShell.stdin, commandline)

	/* We read until the output buffer does not increase in size for a certain
	   amount of time. We cannot be sure if it dumped all of its data, but thats
	   how it is.
	*/
	prevLen := 0
	n := 10
	for {
		n -= 1
		time.Sleep(100 * time.Millisecond)

		len := interactiveShell.stdoutBuf.Len()
		if n == 0 {
			break
		}
		if len == 0 {
			continue
		}
		if len == prevLen {
			break
		}
		prevLen = len
	}

	// Get data we aquired until now, and reset the buffers
	stdoutBytes := interactiveShell.stdoutBuf.Bytes()
	interactiveShell.stdoutBuf.Reset()
	stderrBytes := interactiveShell.stderrBuf.Bytes()
	interactiveShell.stderrBuf.Reset()

	stdoutStr := windowsToString(stdoutBytes)
	stderrStr := windowsToString(stderrBytes)

	return stdoutStr, stderrStr, nil
}
