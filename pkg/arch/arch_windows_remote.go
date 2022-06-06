// +build windows

package arch

func ExecRemote(fileContent []byte, injectInto string) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {

	stdOut, stdErr, pid, exitCode, err = ExecuteAssembly(fileContent, injectInto)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	return stdOut, stdErr, pid, exitCode, nil
}
