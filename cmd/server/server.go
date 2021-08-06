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
	flagDbReadOnly := flag.Bool("dbReadOnly", false, "Only load DB, dont write / update (dont touch DB files)")
	flagDbWriteOnly := flag.Bool("dbWriteOnly", false, "Only write in DB, dont load it on start (overwrite)")
	flag.Parse()

	fmt.Println("Antnium 0.1")
	s := server.NewServer(*flagServerAddr)

	if !*flagDbWriteOnly {
		fmt.Println("Load DB")
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
	}

	if !*flagDbReadOnly {
		fmt.Println("DB Dumper")
		// start DB backups
		go s.PeriodicDbDump()
	}

	s.Serve()
}
