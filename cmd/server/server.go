package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dobin/antnium/pkg/server"
	log "github.com/sirupsen/logrus"
)

func cleanup(s *server.Server) {
	s.DumpDbPackets()
	s.DumpDbClients()
}

func main() {
	flagListenAddr := flag.String("listenaddr", "0.0.0.0:8080", "Server listen address")
	flagDbReadOnly := flag.Bool("dbReadOnly", false, "Only load DB, dont write / update (dont touch DB files)")
	flagDbWriteOnly := flag.Bool("dbWriteOnly", false, "Only write in DB, dont load it on start (overwrite)")
	flag.Parse()

	fmt.Println("Antnium 0.1")
	s := server.NewServer(*flagListenAddr)

	lvl := "debug"
	ll, err := log.ParseLevel(lvl)
	if err != nil {
		ll = log.DebugLevel
	}
	// set global log level
	log.SetLevel(ll)

	// Check prerequisites
	if _, err := os.Stat("./static/"); os.IsNotExist(err) {
		log.Errorf("Server: Could not find required directory: %s", "./static/")
		return
	}
	if _, err := os.Stat("./upload/"); os.IsNotExist(err) {
		log.Errorf("Server: Could not find required directory: %s", "./upload/")
		return
	}

	if !*flagDbWriteOnly {
		// catch ctrl-c so we can save the DB
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cleanup(&s)
			os.Exit(1)
		}()

		// Load DB if any
		err := s.DbLoad()
		if err != nil {
			log.Errorf("Server: Loading DB: %s\n", err.Error())
		}
	}

	if !*flagDbReadOnly {
		fmt.Println("Periodic DB dump enabled")
		// Test DB dump
		err := s.DumpDbClients()
		if err != nil {
			log.Errorf("Server: Could not write DB file in current directory, write access? %s", err.Error())
			return
		}

		// start DB backups
		go s.PeriodicDbDump()
	}

	s.Serve()
}
