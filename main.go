package main

import (
	"flag"
	"fmt"

	"github.com/dobin/antnium/client"
	"github.com/dobin/antnium/server"
)

func main() {
	flagServer := flag.Bool("server", false, "IsServer")
	flagClient := flag.Bool("client", false, "IsClient")
	flag.Parse()

	fmt.Println("Antnium 0.1")

	if *flagServer {
		s := server.NewServer()
		s.Serve()
	}
	if *flagClient {
		c := client.NewClient()
		c.Start()
	}
}
