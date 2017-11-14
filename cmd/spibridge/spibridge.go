package main

import (
	"log"
	"net"
	"github.com/usedbytes/bot_matrix/datalink/spiconn"
	"github.com/usedbytes/bot_matrix/datalink/rpcconn"
)

var addr string = ":9000"

func main() {
	c, err := spiconn.NewSPIConn("/dev/spidev0.0")
	if err != nil {
		panic(err)
	}

	srv, err := rpcconn.NewRPCServ(c)
	if err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	log.Printf("Listening on %s...\n", addr)

	srv.Serve(l)
}
