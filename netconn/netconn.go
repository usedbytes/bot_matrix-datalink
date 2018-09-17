// Copyright 2018 Brian Starkey <stark3y@gmail.com>
package netconn

import (
	"bytes"
	"bufio"
	"encoding/binary"
	"fmt"
	"net"

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

	read chan readResp
}

type readResp struct {
	pkt datalink.Packet
	err error
}

func (c *netconn) readerThread() {
	for {
		n, err := c.c.Read(c.buf[c.cursor:])
		if err != nil {
			// TODO: Should the thread exit?
			c.read <-readResp{ err: err }
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
			c.read <-readResp{
				pkt: datalink.Packet{
					Endpoint: uint8(c.current.ptype),
					Data: c.buf,
				},
			}

			c.remaining = 8
			c.buf = make([]byte, 8)
			c.cursor = 0
			c.current = nil
		}
	}
}

func (c *netconn) Transact(packets []datalink.Packet) ([]datalink.Packet, error) {
	// First send (should we spawn a goroutine for this? If so, need to sync)
	buf := bufio.NewWriter(c.c)
	for _, p := range packets {
		if p.Endpoint == 0 && len(p.Data) == 0 {
			continue
		}

		err := binary.Write(buf, binary.LittleEndian, uint32(p.Endpoint))
		if err != nil {
			fmt.Println(err)
		}
		err = binary.Write(buf, binary.LittleEndian, uint32(len(p.Data)))
		if err != nil {
			fmt.Println(err)
		}
		err = binary.Write(buf, binary.LittleEndian, p.Data)
		if err != nil {
			fmt.Println(err)
		}
	}
	buf.Flush()

	// Then receive
	rxPkts := make([]datalink.Packet, 0, 4)
	for {
		select {
		case resp := <-c.read:
			if resp.err != nil {
				return rxPkts, resp.err
			}
			rxPkts = append(rxPkts, resp.pkt)
		default:
			return rxPkts, nil
		}
	}
}

func NewNetconn(c net.Conn) datalink.Transactor {
	conn := &netconn{
		c: c,
		buf: make([]byte, 8),
		remaining: 8,
		read: make(chan readResp, 5),
	}

	go conn.readerThread()

	return conn
}
