package inject

import "unsafe"

func UUIDFromStringA(uuidString string, uuid uintptr) (uintptr, error) {
	status, _, err := uuidFromStringA.Call(uintptr(unsafe.Pointer(StringToCharPtr(uuidString))), uuid)
	return status, err
}
