// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package datalink

import (
	"log"
)

type Packet struct {
	Endpoint uint8
	Data []byte
}

type Transactor interface {
	Transact([]Packet) ([]Packet, error)
}

func PumpTransactor(conn Transactor, tx <-chan Packet,
		    rx chan<- Packet, stop <-chan bool,
		    ticker <-chan struct{}) {
	// XXX: try and adjust minNum by heuristics to get the necessary
	// throughput based on actual utilisation
	minNum := 4
	toSend := make([]Packet, 0, minNum)

	for {
		select {
		case _ = <-ticker:
			if len(toSend) < minNum {
				toSend = append(toSend, make([]Packet, minNum - len(toSend))...)
			}

			pkts, err := conn.Transact(toSend)
			if err != nil {
				// XXX: Signal this error back to caller
				log.Printf("Error! %s\n", err)
			} else {
				for _, p := range pkts {
					rx <-p
				}
			}

			toSend = make([]Packet, 0, minNum)

		case p := <-tx:
			toSend = append(toSend, p)

		case _ = <-stop:
			return
		}
	}
}
