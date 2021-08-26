package main

import (
	"fmt"

	"github.com/dobin/antnium/pkg/downstreamclient"
)

func main() {
	fmt.Println("Antnium 0.1")

	downstreamClient := downstreamclient.MakeDownstreamClient()
	downstreamClient.StartClient("")
}
