package main

import (
	"flag"
	"fmt"

	"github.com/dobin/antnium/pkg/server"
)

func main() {
	flagServerAddr := flag.String("serveraddr", "127.0.0.1:4444", "Server listen address")
	flag.Parse()

	fmt.Println("Antnium 0.1")

	s := server.NewServer(*flagServerAddr)
	s.Serve()
}
