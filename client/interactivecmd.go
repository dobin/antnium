package client

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

type InteractiveCmd struct {
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	stdoutBuf *bytes.Buffer
	stderrBuf *bytes.Buffer
}

func makeInteractiveCmd() InteractiveCmd {
	interactiveCmd := InteractiveCmd{
		nil,
		nil,
		nil,
		nil,
		nil,
	}
	return interactiveCmd
}

func (interactiveCmd *InteractiveCmd) open() (string, string, error) {
	// Setup command
	cmd := exec.Command("cmd", "/a")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", "", err
	}

	// Start it
	if err := cmd.Start(); err != nil {
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

	interactiveCmd.stdin = stdin
	interactiveCmd.stdout = stdout
	interactiveCmd.stderr = stderr
	interactiveCmd.stdoutBuf = bytes.NewBuffer(nil)
	interactiveCmd.stderrBuf = bytes.NewBuffer(nil)

	// read the stdout continuously in a separate goroutine and capture it in our vars
	// Read() will block if no data is available.
	go func() {
		for {
			part := make([]byte, 128)
			n, err := interactiveCmd.stdout.Read(part)
			if err != nil {
				break
			}
			interactiveCmd.stdoutBuf.Write(part[0:n])
		}
	}()
	go func() {
		for {
			part := make([]byte, 128)
			n, err := interactiveCmd.stderr.Read(part)
			if err != nil {
				break
			}
			interactiveCmd.stderrBuf.Write(part[0:n])
		}
	}()

	return string(stdoutCut1) + string(stdoutCut2), string(""), nil
}

func (interactiveCmd *InteractiveCmd) issue(cmd string) (string, string) {
	// Give command to cmd
	fmt.Fprintln(interactiveCmd.stdin, cmd)

	/* We read until the output buffer does not increase in size for a certain
	   amount of time. We cannot be sure if it dumped all of its data, but thats
	   how it is.
	*/
	prevLen := 0
	for {
		time.Sleep(100 * time.Millisecond)

		len := interactiveCmd.stdoutBuf.Len()
		if len == 0 {
			continue
		}
		if len == prevLen {
			break
		}
		prevLen = len
	}

	// Get data we aquired until now, and reset the buffers
	stdout := interactiveCmd.stdoutBuf.String()
	interactiveCmd.stdoutBuf.Reset()
	stderr := interactiveCmd.stderrBuf.String()
	interactiveCmd.stderrBuf.Reset()

	return stdout, stderr
}
