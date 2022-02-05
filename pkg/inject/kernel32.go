package inject

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func CreateThread(startAddress uintptr) uintptr {

	thread, _, _ := createThread.Call(0, 0, startAddress, uintptr(0), 0, 0)
	return thread
}

func VirtualAlloc(address uintptr, size int, allocationType uint64, protect uint64) uintptr {

	addr, _, _ := virtualAlloc.Call(address, uintptr(size), windows.MEM_COMMIT|windows.MEM_RESERVE, windows.PAGE_READWRITE)
	return addr
}

func HeapCreate(options uint32, initialSize int, maximumSize uint32) uintptr {

	heap, _, _ := heapCreate.Call(uintptr(options), uintptr(initialSize), uintptr(maximumSize))
	return heap
}

func HeapAlloc(heap uintptr, dwFlags uint32, dwBytes int) uintptr {

	allocatedMemory, _, _ := heapAlloc.Call(heap, uintptr(dwFlags), uintptr(dwBytes))
	return allocatedMemory
}

func OpenProcess(desiredAccess uint32, inheritHandle uint32, processId uint32) (uintptr, error) {

	pHandle, _, err := openProcess.Call(uintptr(desiredAccess), uintptr(inheritHandle), uintptr(processId))
	return pHandle, err

}

func VirtualAllocEx(process uintptr, address uintptr, length int, allocationType uint32, protect uint32) uintptr {

	memptr, _, _ := virtualAllocEx.Call(process, uintptr(address), uintptr(length), uintptr(allocationType), uintptr(protect))
	return memptr
}

func VirtualAllocEx2(process uintptr, address uintptr, length uintptr, allocationType uint32, protect uint32) uintptr {

	memptr, _, _ := virtualAllocEx.Call(process, uintptr(address), uintptr(length), uintptr(allocationType), uintptr(protect))
	return memptr
}

func WriteProcessMemory(process uintptr, baseAddress uintptr, buffer []byte) uint32 {

	var nbytes uint32
	writeProcessMemory.Call(uintptr(process), baseAddress, (uintptr)(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)), uintptr(nbytes))
	return nbytes
}

func WriteProcessMemory2(process uintptr, baseAddress uintptr, buffer []byte, length uint32) (uint32, error) {

	var nbytes uint32
	_, _, err := writeProcessMemory.Call(process, baseAddress, (uintptr(unsafe.Pointer(&buffer[0]))), uintptr(length), uintptr(unsafe.Pointer(&nbytes)))
	return nbytes, err
}

func CreateRemoteThread(process uintptr, threadAttributes uintptr, stackSize uint64, startAddress uintptr, paramater uintptr, creationFlags uint32, threadID uint32) {

	createRemoteThread.Call(uintptr(process), threadAttributes, uintptr(stackSize), startAddress, uintptr(paramater), uintptr(creationFlags), uintptr(threadID))
	return
}

func CloseHandle(handle uintptr) {

	closeHandle.Call(handle)
	return
}

func IsWow64Process(handle uintptr) uint32 {
	var bitness uint32
	isWow64Process.Call(handle, uintptr(unsafe.Pointer(&bitness)))
	return bitness
}

func VirtualProtect(address uintptr, size int, newProtect uint32) uint32 {

	var oldProtect uint32
	virtualProtect.Call(address, uintptr(size), uintptr(newProtect), (uintptr)(unsafe.Pointer(&oldProtect)))
	return oldProtect
}

func WaitForSingleObject(thread uintptr, milliseconds uint32) {

	waitForSingleObject.Call(uintptr(windows.Handle(thread)), uintptr(milliseconds))
	return
}

func GetProcAddress(module uintptr, procName string) uintptr {

	address, _, _ := getProcAddress.Call(module, uintptr(unsafe.Pointer(StringToCharPtr(procName))))
	return address
}

func GetModuleHandleA(moduleName string) uintptr {

	handle, _, _ := getModuleHandleA.Call(uintptr(unsafe.Pointer(StringToCharPtr(moduleName))))
	return handle
}

func CreateToolhelp32Snapshot(flags uint32, pid uint32) uintptr {
	handle, _, _ := createToolhelp32Snapshot.Call(uintptr(flags), uintptr(pid))
	return handle
}

func Process32First(snapshot uintptr, processEntry *windows.ProcessEntry32) (uintptr, error) {

	result, _, err := process32First.Call(snapshot, (uintptr)(unsafe.Pointer(processEntry)))
	return result, err
}

func Process32Next(snapshot uintptr, processEntry *windows.ProcessEntry32) (uintptr, error) {

	result, _, err := process32Next.Call(snapshot, (uintptr)(unsafe.Pointer(processEntry)))
	return result, err
}

func Thread32First(snapshot uintptr, threadEntry *windows.ThreadEntry32) (uintptr, error) {

	result, _, err := thread32First.Call(snapshot, (uintptr)(unsafe.Pointer(threadEntry)))
	return result, err
}

func Thread32Next(snapshot uintptr, threadEntry *windows.ThreadEntry32) (uintptr, error) {

	result, _, err := thread32Next.Call(snapshot, (uintptr)(unsafe.Pointer(threadEntry)))
	return result, err
}

func OpenThread(desiredAccess uint32, inheritHandle uint32, threadId uint32) (uintptr, error) {

	tHandle, _, err := openThread.Call(uintptr(desiredAccess), uintptr(inheritHandle), uintptr(threadId))
	return tHandle, err
}

func QueueUserAPC(pfnAPC *uintptr, tHandle uintptr) uint32 {
	result, _, _ := queueUserAPC.Call((uintptr)(unsafe.Pointer(&pfnAPC)), tHandle, 0)
	return uint32(result)
}

func EnumSystemLocalesA(lpLocaleEnumProc uintptr, dwFlags uint32) error {
	_, _, err := enumSystemLocalesA.Call(lpLocaleEnumProc, uintptr(dwFlags))
	return err
}

func GetCurrentThreadId() (uint32, error) {
	result, _, err := getCurrentThreadId.Call()
	return uint32(result), err
}

func SetConsoleCtrlHandler(handlerRoutine HANDLER, add uint32) error {

	_, _, err := setConsoleCtrlHandler.Call(uintptr(syscall.NewCallback(handlerRoutine)), uintptr(add))
	return err
}

func CreateProcessA(appName string,
	commandLine string,
	processAttributes uintptr,
	threadAttributes uintptr,
	inheritHandles uint32,
	creationFlags uint32,
	env uintptr,
	currentDir uintptr) (uint32, *syscall.StartupInfo, *syscall.ProcessInformation, error) {

	SI := new(syscall.StartupInfo)
	PI := new(syscall.ProcessInformation)

	result, _, err := createProcessA.Call(uintptr(unsafe.Pointer(StringToCharPtr(appName))),
		uintptr(unsafe.Pointer(StringToCharPtr(commandLine))),
		processAttributes,
		threadAttributes,
		uintptr(inheritHandles),
		uintptr(creationFlags),
		env,
		currentDir,
		uintptr(unsafe.Pointer(SI)),
		uintptr(unsafe.Pointer(PI)),
	)
	return uint32(result), SI, PI, err
}

// https://github.com/abdullah2993/go-runpe/blob/403894fc2c3152c1f8ac98221250f1d46fd70bff/runpe.go#L288
func GetThreadContext(hThread uintptr, ctx *CONTEXT) error {
	_, _, err := getThreadContext.Call(hThread, (uintptr)(unsafe.Pointer(ctx)))
	return err
}

func ReadProcessMemory(process uintptr, baseAddress uintptr, buffer []byte, size uint32) (uint32, error) {

	var nbytes uint32
	_, _, err := readProcessMemory.Call(uintptr(process), baseAddress, (uintptr(unsafe.Pointer(&buffer[0]))), uintptr(size), uintptr(unsafe.Pointer(&nbytes)))
	return nbytes, err
}

func SetThreadContext(hThread uintptr, ctx CONTEXT) error {
	_, _, err := setThreadContext.Call(hThread, uintptr(unsafe.Pointer(&ctx)))
	return err
}

func ResumeThread(hThread uintptr) error {
	_, _, err := resumeThread.Call(hThread)
	return err
}

func LoadLibraryA(LibFileName string) (uintptr, error) {
	handle, _, err := loadLibraryA.Call(uintptr(unsafe.Pointer(StringToCharPtr(LibFileName))))
	return handle, err
}
