package model

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"golang.org/x/sys/windows"
)

type DirEntry struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Mode     string    `json:"mode"`
	Modified time.Time `json:"modified"`
	IsDir    bool      `json:"isDir"`
}

func ListDirectory(path string) ([]DirEntry, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	dirList := make([]DirEntry, 0)
	for _, file := range files {
		dl := DirEntry{
			file.Name(),
			file.Size(),
			"", // Mode()
			file.ModTime(),
			file.IsDir(),
		}
		dirList = append(dirList, dl)
	}

	return dirList, err
}

func MakePacketArgumentFrom(packetArgument PacketArgument) (string, []string, error) {
	args := make([]string, 0)

	executable, ok := packetArgument["executable"]
	if !ok {
		return "", nil, fmt.Errorf("No executable given")
	}

	n := 0
	for {
		nr := strconv.Itoa(n)
		key := "param" + nr
		_, ok := packetArgument[key]
		if ok {
			args = append(args, packetArgument[key])
		} else {
			break
		}
		n = n + 1
	}

	return executable, args, nil
}

// https://coolaj86.com/articles/golang-and-windows-and-admins-oh-my/
func GetPermissions() (bool, bool, error) {
	var sid *windows.SID

	// Although this looks scary, it is directly copied from the
	// official windows documentation. The Go API for this is a
	// direct wrap around the official C++ API.
	// See https://docs.microsoft.com/en-us/windows/desktop/api/securitybaseapi/nf-securitybaseapi-checktokenmembership
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		return false, false, err
	}

	// This appears to cast a null pointer so I'm not sure why this
	// works, but this guy says it does and it Works for Me™:
	// https://github.com/golang/go/issues/28804#issuecomment-438838144
	token := windows.Token(0)

	member, err := token.IsMember(sid)
	if err != nil {
		return false, false, err
	}

	// Also note that an admin is _not_ necessarily considered
	// elevated.
	// For elevation see https://github.com/mozey/run-as-admin

	isElevated := token.IsElevated()
	isAdmin := member
	return isElevated, isAdmin, nil
}
