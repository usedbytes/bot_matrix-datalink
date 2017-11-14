// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package datalink

type Packet struct {
	Endpoint uint8
	Data []byte
}

type Transactor interface {
	Transact([]Packet) ([]Packet, error)
}

type Protocol interface {
	Serialise([]Packet) []byte
	DeSerialise([]byte) ([]Packet, error)
}

type Transport interface {
	Transfer([]byte) ([]byte, error)
}

type Connection struct {
	protocol Protocol
	transport Transport
}

func NewConnection(proto Protocol, xport Transport) *Connection {
	return &Connection{ proto, xport }
}

func (c *Connection) Transact(packets []Packet) ([]Packet, error) {
	tx := c.protocol.Serialise(packets)

	rx, err := c.transport.Transfer(tx)
	if err != nil {
		return nil, err
	}

	rxPkts, err := c.protocol.DeSerialise(rx)
	if err != nil {
		return nil, err
	}

	return rxPkts, nil
}
