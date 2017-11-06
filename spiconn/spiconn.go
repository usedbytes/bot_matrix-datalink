// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package spiconn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"github.com/sigurn/crc8"
	"github.com/usedbytes/bot_matrix/datalink/connection"
	"github.com/usedbytes/bot_matrix/datalink/packet"
)

/*
struct spi_pl_packet {
	uint8_t id;
	uint8_t type;
	uint8_t nparts;
	uint8_t flags;
	uint8_t data[SPI_PACKET_DATA_LEN];
	uint8_t crc;
};
*/

type spiXport struct {

}

func (x *spiXport) Transfer(data []byte) ([]byte, error) {
	log.Printf("SPI Xfer: %x\n", data)

	return data, nil
}

type spiProto struct {
	id uint8
	datalen int
	crc *crc8.Table
}

func writeBuf(buf *bytes.Buffer, v interface{}) error {
	return binary.Write(buf, binary.LittleEndian, v)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (p *spiProto) serialise(into *bytes.Buffer, pkt packet.Packet) {
	nparts := uint8((len(pkt.Data) + p.datalen - 1) / p.datalen)

	for i, start := 0, 0; i < len(pkt.Data); i += p.datalen {
		p.id += 1
		// nparts is actually "number of parts still to come"
		// so decrement it before anything
		nparts -= 1

		hdr := []byte{ p.id, pkt.Endpoint, nparts, byte(0) }
		writeBuf(into, hdr)

		end := min(len(pkt.Data), i + p.datalen)
		writeBuf(into, pkt.Data[i:end])
		if end - i < p.datalen {
			// If the last packet is short, we've got to pad it with
			// zeroes.
			writeBuf(into, make([]byte, p.datalen - (end - i)))
		}

		crc := crc8.Checksum(into.Bytes()[start:into.Len()], p.crc)
		writeBuf(into, crc)

		start = into.Len()
	}
}

func (p *spiProto) Serialise(pkts []packet.Packet) []byte {
	buf := new(bytes.Buffer)

	for _, pkt := range pkts {
		p.serialise(buf, pkt)
	}

	return buf.Bytes()
}

func (p *spiProto) DeSerialise(data []byte) ([]packet.Packet, error) {
	hdrLen := 4
	packetLen := p.datalen + hdrLen + 1

	pkts := make([]packet.Packet, 0, len(data) / packetLen)

	var payload []byte
	var id, ep, nparts byte

	for i := 0; i < len(data); {
		if len(data) < i + packetLen {
			return pkts, fmt.Errorf("Short data. Have %d bytes, need %d",
						len(data), i + packetLen)
		}

		crc := crc8.Checksum(data[i:i + packetLen - 1], p.crc)
		if crc != data[i + packetLen - 1] {
			return pkts, fmt.Errorf("CRC error in packet %d", len(pkts) + 1)
		}

		if payload == nil {
			payload = make([]byte, 0, int(data[i + 2]) * p.datalen)
		} else {
			if data[i] != id + 1 {
				return pkts, fmt.Errorf("Invalid packet ID. Expected %d got %d", id + 1, data[i])
			}

			if data[i + 1] != ep {
				return pkts, fmt.Errorf("Invalid Endpoint. Expected %d got %d", ep, data[i + 1])
			}

			if data[i + 2] != nparts - 1 {
				return pkts, fmt.Errorf("Invalid nparts. Expected %d got %d", nparts - 1, data[i + 2])
			}
		}

		id = data[i]
		ep = data[i + 1]
		nparts = data[i + 2]
		payload = append(payload, data[i + hdrLen:i + packetLen - 1]...)

		if nparts == 0 {
			pkts = append(pkts, packet.Packet{
				Endpoint: ep,
				Data: payload,
			})
			payload = nil
		}

		i += packetLen
	}

	return pkts, nil
}

func NewSPIConn(device string) (*connection.Connection, error) {
	proto := spiProto{
		id: 0,
		datalen: 4,
		crc: crc8.MakeTable(crc8.CRC8),
	}

	_ = device

	conn := connection.NewConnection(&proto, &spiXport{})

	return conn, nil
}
