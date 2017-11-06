package main

import (
	"log"
	"github.com/usedbytes/bot_matrix/datalink/packet"
	"github.com/usedbytes/bot_matrix/datalink/spiconn"
)

func main() {
	c, _ := spiconn.NewSPIConn("/dev/spidev0.0")

	p := packet.Packet{
		Endpoint: 1,
		Data: []byte{10, 11, 12, 13, 14, 15},
	}

	s, err := c.Transact([]packet.Packet{p})
	if err != nil {
		log.Println(err)
	}

	log.Printf("%#v\n", s)
}
