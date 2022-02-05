package inject

import (
	"golang.org/x/sys/windows"
)

var (
	ntdll    = windows.NewLazySystemDLL("ntdll.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	psapi    = windows.NewLazySystemDLL("psapi.dll")
	rpcrt4   = windows.NewLazySystemDLL("Rpcrt4.dll")
	user32   = windows.NewLazySystemDLL("User32.dll")

	// NTDLL
	rtlCopyMemory        = ntdll.NewProc("RtlCopyMemory")
	rtlMoveMemory        = ntdll.NewProc("RtlMoveMemory")
	ntUnmapViewOfSection = ntdll.NewProc("NtUnmapViewOfSection")

	// KERNEL32
	createThread             = kernel32.NewProc("CreateThread")
	virtualAlloc             = kernel32.NewProc("VirtualAlloc")
	heapCreate               = kernel32.NewProc("HeapCreate")
	heapAlloc                = kernel32.NewProc("HeapAlloc")
	openProcess              = kernel32.NewProc("OpenProcess")
	virtualAllocEx           = kernel32.NewProc("VirtualAllocEx")
	virtualProtect           = kernel32.NewProc("VirtualProtect")
	writeProcessMemory       = kernel32.NewProc("WriteProcessMemory")
	createRemoteThread       = kernel32.NewProc("CreateRemoteThread")
	closeHandle              = kernel32.NewProc("CloseHandle")
	isWow64Process           = kernel32.NewProc("IsWow64Process")
	waitForSingleObject      = kernel32.NewProc("WaitForSingleObject")
	getProcAddress           = kernel32.NewProc("GetProcAddress")
	getModuleHandleA         = kernel32.NewProc("GetModuleHandleA")
	createToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	process32First           = kernel32.NewProc("Process32First")
	process32Next            = kernel32.NewProc("Process32Next")
	thread32First            = kernel32.NewProc("Thread32First")
	thread32Next             = kernel32.NewProc("Thread32Next")
	openThread               = kernel32.NewProc("OpenThread")
	queueUserAPC             = kernel32.NewProc("QueueUserAPC")
	enumSystemLocalesA       = kernel32.NewProc("EnumSystemLocalesA")
	getCurrentThreadId       = kernel32.NewProc("GetCurrentThreadId")
	setConsoleCtrlHandler    = kernel32.NewProc("SetConsoleCtrlHandler")
	createProcessA           = kernel32.NewProc("CreateProcessA")
	getThreadContext         = kernel32.NewProc("GetThreadContext")
	readProcessMemory        = kernel32.NewProc("ReadProcessMemory")
	setThreadContext         = kernel32.NewProc("SetThreadContext")
	resumeThread             = kernel32.NewProc("ResumeThread")
	loadLibraryA             = kernel32.NewProc("LoadLibraryA")

	// PSAPI
	enumProcesses = psapi.NewProc("EnumProcesses")

	// rpcrt4
	uuidFromStringA = rpcrt4.NewProc("UuidFromStringA")

	// user32
	setWindowsHookExA        = user32.NewProc("SetWindowsHookExA")
	getMessageW              = user32.NewProc("GetMessageW")
	translateMessage         = user32.NewProc("TranslateMessage")
	dispatchMessage          = user32.NewProc("DispatchMessage")
	unhookWindowsHookEx      = user32.NewProc("UnhookWindowsHookEx")
	postThreadMessage        = user32.NewProc("PostThreadMessage")
	callNextHookEx           = user32.NewProc("CallNextHookEx")
	findWindowA              = user32.NewProc("FindWindowA")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
)
