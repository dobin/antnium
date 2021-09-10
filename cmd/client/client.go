package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dobin/antnium/pkg/client"
	log "github.com/sirupsen/logrus"
)

func main() {
	flagServerUrl := flag.String("addr", "", "ENV: PROXY")
	flagProxyUrl := flag.String("proxy", "", "ENV: ADDR")
	flag.Parse()

	fmt.Println("Antnium 0.1")
	lvl := "debug"
	ll, err := log.ParseLevel(lvl)
	if err != nil {
		ll = log.DebugLevel
	}
	// set global log level
	log.SetLevel(ll)

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
