package inject

import "unsafe"

func RtlCopyMemory(destination uintptr, source []byte) {
	rtlCopyMemory.Call(destination, (uintptr)(unsafe.Pointer(&source[0])), uintptr(len(source)))
	return
}

func RtlMoveMemory(source uintptr, length int) int {
	var destination int
	rtlMoveMemory.Call((uintptr)(unsafe.Pointer(&destination)), source, uintptr(length))
	return destination
}

func RtlMoveMemory2(destination uintptr, source []byte) {
	rtlMoveMemory.Call(destination, (uintptr)(unsafe.Pointer(&source[0])), uintptr(len(source)))
}

func NtUnmapViewOfSection(processHandle uintptr, baseAddress uintptr) (uintptr, error) {
	r, _, err := ntUnmapViewOfSection.Call(processHandle, baseAddress)
	return r, err
}
