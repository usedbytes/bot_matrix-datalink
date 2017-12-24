// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package datalink

type Packet struct {
	Endpoint uint8
	Data []byte
}

type Transactor interface {
	Transact([]Packet) ([]Packet, error)
}
