package main

import (
	"fmt"

	"github.com/dobin/antnium/pkg/wingman"
)

func main() {
	fmt.Println("Antnium 0.1")

	wingman := wingman.MakeWingman()
	wingman.StartWingman("")
}
