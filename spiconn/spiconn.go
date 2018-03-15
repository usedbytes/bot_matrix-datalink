// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package spiconn

import (
	"fmt"
	"encoding/hex"
	"github.com/ecc1/spi"
	"github.com/sigurn/crc8"
	"github.com/usedbytes/bot_matrix/datalink"
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

type transferrer interface {
	transfer(data []byte) ([]byte, error)
	transferMultiple(data [][]byte) ([][]byte, error)
}

type spiXferrer struct {
	dev *spi.Device
}

func (s *spiXferrer) transfer(data []byte) ([]byte, error) {
	tmp := make([]byte, len(data))
	copy(tmp, data)

	err := s.dev.Transfer(tmp)

	return tmp, err
}

func (s *spiXferrer) transferMultiple(data [][]byte) ([][]byte, error) {
	ret := make([][]byte, len(data))

	for i, t := range data {
		rx, err := s.transfer(t)
		if err != nil {
			return nil, err
		}
		ret[i] = rx
	}

	return ret, nil
}

type deSerCtx struct {
	id, ep, nparts byte
	payload []byte
}

type spiLink struct {
	xferrer transferrer

	id uint8
	datalen int
	crc *crc8.Table

	ctx deSerCtx
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *spiLink) serialiseOne(pkt datalink.Packet) [][]byte {
	// Slightly ugly way to handle empty packets.
	if pkt.Data == nil || len(pkt.Data) == 0 {
		pkt.Data = make([]byte, 1)
	}

	nparts := uint8((len(pkt.Data) + s.datalen - 1) / s.datalen)
	ret := make([][]byte, 0, nparts)

	for i := 0; i < len(pkt.Data); i += s.datalen {
		// nparts is actually "number of parts still to come"
		// so decrement it before anything
		nparts -= 1
		s.id += 1

		d := make([]byte, 4 + s.datalen + 1)
		d[0] = s.id
		d[1] = pkt.Endpoint
		d[2] = nparts
		d[3] = 0

		end := min(len(pkt.Data), i + s.datalen)
		copy(d[4:], pkt.Data[i:end])

		d[len(d)-1] = crc8.Checksum(d[:len(d)-1], s.crc)

		ret = append(ret, d)
	}

	return ret
}

func (s *spiLink) serialise(packets []datalink.Packet) [][]byte {
	// len(packets) is a good approximation - but if some packets are longer
	// than one transfer, then it will be wrong.
	transfers := make([][]byte, 0, len(packets))

	for _, pkt := range packets {
		transfers = append(transfers, s.serialiseOne(pkt)...)
	}

	return transfers
}

func dumpTransfers(data [][]byte) {
	for _, t := range data {
		fmt.Println(hex.Dump(t))
	}
}

func nextID(id uint8) uint8 {
	id = id + 1
	if (id >= 0x80) {
		id = 0
	}
	return id
}

func (s *spiLink) deSerialise(data [][]byte) ([]datalink.Packet, error) {
	hdrLen := 4
	packetLen := s.datalen + hdrLen + 1

	pkts := make([]datalink.Packet, 0, len(data))

	for i := 0; i < len(data); i++ {
		transfer := data[i]

		if len(transfer) < packetLen {
			s.ctx.payload = nil
			return pkts, fmt.Errorf("Short data. Have %d bytes, need %d",
						len(transfer), packetLen)
		}

		if crc8.Checksum(transfer, s.crc) != 0 {
			s.ctx.payload = nil
			return pkts, fmt.Errorf("CRC error in transfer %d.", i)
		}

		if s.ctx.payload == nil {
			s.ctx.payload = make([]byte, 0, int(transfer[2]) * s.datalen)
		} else {
			if transfer[0] != nextID(s.ctx.id) {
				s.ctx.payload = nil
				return pkts, fmt.Errorf("Invalid packet ID. Expected %d got %d", s.ctx.id + 1, transfer[0])
			}

			if transfer[1] == 0 {
				s.ctx.id = transfer[0]
				continue
			}

			if transfer[1] != s.ctx.ep {
				s.ctx.payload = nil
				return pkts, fmt.Errorf("Invalid Endpoint. Expected %d got %d", s.ctx.ep, transfer[1])
			}

			if transfer[2] != s.ctx.nparts - 1 {
				s.ctx.payload = nil
				return pkts, fmt.Errorf("Invalid nparts. Expected %d got %d", s.ctx.nparts - 1, transfer[2])
			}
		}

		s.ctx.id = transfer[0]
		s.ctx.ep = transfer[1]
		s.ctx.nparts = transfer[2]
		s.ctx.payload = append(s.ctx.payload, transfer[hdrLen:packetLen - 1]...)

		if s.ctx.nparts == 0 {
			pkts = append(pkts, datalink.Packet{
				Endpoint: s.ctx.ep,
				Data: s.ctx.payload,
			})
			s.ctx.payload = nil
		}
	}

	return pkts, nil
}

func (s *spiLink) Transact(packets []datalink.Packet) ([]datalink.Packet, error) {
	transfers := s.serialise(packets)

	rx, err := s.xferrer.transferMultiple(transfers)
	if err != nil {
		return nil, err
	}

	rxPkts, err := s.deSerialise(rx)
	if err != nil {
		return nil, err
	}

	return rxPkts, nil
}

func NewSPIConn(device string) (datalink.Transactor, error) {
	link := &spiLink{
		id: 0,
		datalen: 32,
		crc: crc8.MakeTable(crc8.CRC8),
	}

	dev, err := spi.Open(device, 1000000, 0)
	if err != nil {
		return nil, err
	}

	link.xferrer = &spiXferrer{ dev, }

	return link, nil
}
