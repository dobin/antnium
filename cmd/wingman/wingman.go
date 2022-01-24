package main

import (
	"C"
	"fmt"
)
import "github.com/dobin/antnium/pkg/wingman"

/*
In a admin shell:
C:\Windows\System32\rundll32.exe .\wingman.dll,Start
*/

//export Start
func Start() {
	wingman := wingman.MakeWingman()
	wingman.StartWingman("")
}

func main() {
	fmt.Println("Antnium 0.1")

	wingman := wingman.MakeWingman()
	wingman.StartWingman("directory")
}
