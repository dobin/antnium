package inject

import "unsafe"

func EnumProcesses() ([1024]uint32, uint32) {
	var processes [1024]uint32
	var cbNeeded uint32
	enumProcesses.Call((uintptr)(unsafe.Pointer(&processes)), uintptr(len(processes)), (uintptr)(unsafe.Pointer(&cbNeeded)))
	return processes, cbNeeded
}
