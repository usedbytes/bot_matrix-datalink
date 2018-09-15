// Copyright 2018 Brian Starkey <stark3y@gmail.com>
package netconn

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/usedbytes/bot_matrix/datalink"
)

/*
struct netconn_packet {
	uint32_t ptype
	uint32_t length;
	uint8_t data[];
}
*/
type Packet struct {
	ptype uint32
	length uint32
	data []byte
}

type netconn struct {
	c net.Conn

	// TODO: Could wrap this up in some buffer class
	buf []byte
	cursor uint32

	current *Packet
	remaining uint32
}

func (c *netconn) Transact(packets []datalink.Packet) ([]datalink.Packet, error) {
	// First send (should we spawn a goroutine for this? If so, need to sync)
	for _, p := range packets {
		binary.Write(c.c, binary.LittleEndian, uint32(p.Endpoint))
		binary.Write(c.c, binary.LittleEndian, uint32(len(p.Data)))
		binary.Write(c.c, binary.LittleEndian, p.Data)
	}

	// Then receive
	nrx := 0
	rxPkts := make([]datalink.Packet, 0, 4)
	for {
		// TODO: Find a better way for nonblock reads
		c.c.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
		n, err := c.c.Read(c.buf[c.cursor:])
		if err != nil {
			// TODO: Catch errors which aren't timeouts
			break
		} else if n == 0 {
			break
		}

		c.remaining -= uint32(n)
		c.cursor += uint32(n)

		if c.remaining != 0 {
			continue
		}

		if c.current == nil {
			c.current = &Packet{}
			bbuf := bytes.NewBuffer(c.buf)
			binary.Read(bbuf, binary.LittleEndian, &c.current.ptype)
			binary.Read(bbuf, binary.LittleEndian, &c.remaining)

			c.buf = make([]byte, c.remaining)
			c.cursor = 0
		} else {
			rxPkts = append(rxPkts, datalink.Packet{})
			rxPkts[nrx].Endpoint = uint8(c.current.ptype)
			rxPkts[nrx].Data = c.buf
			nrx++

			c.remaining = 8
			c.buf = make([]byte, 8)
			c.cursor = 0
			c.current = nil
		}
	}

	return rxPkts, nil
}

func NewNetconn(c net.Conn) datalink.Transactor {
	conn := &netconn{
		c: c,
		buf: make([]byte, 8),
		remaining: 8,
	}

	return conn
}
