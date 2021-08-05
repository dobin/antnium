package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dobin/antnium/pkg/server"
)

func cleanup(s *server.Server) {
	s.DumpDbPackets()
	s.DumpDbClients()
}

func main() {
	flagServerAddr := flag.String("serveraddr", "127.0.0.1:4444", "Server listen address")
	flag.Parse()

	fmt.Println("Antnium 0.1")

	s := server.NewServer(*flagServerAddr)

	// catch ctrl-c
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(&s)
		os.Exit(1)
	}()

	// Load DB if any
	s.DbLoad()

	// start DB backups
	go s.PeriodicDbDump()

	s.Serve()
}
