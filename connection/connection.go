// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package connection

import (
	"github.com/usedbytes/bot_matrix/datalink/packet"
	"github.com/usedbytes/bot_matrix/datalink/protocol"
	"github.com/usedbytes/bot_matrix/datalink/transport"
)

type Transactor interface {
	Transact([]packet.Packet) ([]packet.Packet, error)
}

type Connection struct {
	protocol protocol.Protocol
	transport transport.Transport
}

func NewConnection(proto protocol.Protocol, xport transport.Transport) *Connection {
	return &Connection{ proto, xport }
}

func (c *Connection) Transact(packets []packet.Packet) ([]packet.Packet, error) {
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
