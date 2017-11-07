package main

import (
	"log"
	"github.com/usedbytes/bot_matrix/datalink/packet"
	"github.com/usedbytes/bot_matrix/datalink/spiconn"
)

func main() {
	c, _ := spiconn.NewSPIConn("/dev/spidev0.0")

	p := packet.Packet{
		Endpoint: 0x37,
		Data: []byte{0x0a, 0x0b, 0x0c, 0x0d, },
	}

	s, err := c.Transact([]packet.Packet{p})
	if err != nil {
		log.Println(err)
	}

	log.Printf("%#v\n", s)
}
