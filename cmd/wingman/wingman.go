package main

import (
	"C"
	"fmt"
)
import (
	"flag"

	"github.com/dobin/antnium/pkg/wingman"
)

/*
In a admin shell:
C:\Windows\System32\rundll32.exe .\wingman.dll,Start
*/

//export Start
func Start() {
	wingman := wingman.MakeWingman()
	wingman.StartWingman("", "")
}

func main() {
	fmt.Println("Wingman 0.1")

	/*
		tcp       localhost:50000
		directory c:\temp\
	*/
	proto := flag.String("proto", "", "proto")
	data := flag.String("data", "", "data")
	flag.Parse()

	wingman := wingman.MakeWingman()
	wingman.StartWingman(*proto, *data)
}
