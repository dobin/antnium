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
	flagAddr := flag.String("listenaddr", "127.0.0.1:4444", "Server listen address")
	flag.Parse()

	fmt.Println("Antnium 0.1")

	if *flagServer {
		s := server.NewServer(*flagAddr)
		s.Serve()
	}
	if *flagClient {
		c := client.NewClient()
		c.Start()
	}
}
