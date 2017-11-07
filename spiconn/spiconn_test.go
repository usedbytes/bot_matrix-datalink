// Copyright 2017 Brian Starkey <stark3y@gmail.com>

package spiconn

import (
	"bytes"
	"testing"

	"github.com/sigurn/crc8"
	"github.com/usedbytes/bot_matrix/datalink/packet"
)

func TestInnerSerialise(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	buf := new(bytes.Buffer)
	pkt := packet.Packet{
		Endpoint: 0x37,
		Data:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
	}
	expect := []byte{0x01, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d, 0xdd}

	proto.serialise(buf, pkt)
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, buf.Bytes())
	}
}

func TestInnerSerialiseShortData(t *testing.T) {
	proto := &spiProto{
		id:      1,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	buf := new(bytes.Buffer)
	pkt := packet.Packet{
		Endpoint: 0x37,
		Data:     []byte{0x0a},
	}

	buf.Reset()
	expect := []byte{0x02, 0x37, 0x00, 0x00, 0x0a, 0x00, 0x00, 0x00}
	expect = append(expect, crc8.Checksum(expect, proto.crc))
	proto.serialise(buf, pkt)
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, buf.Bytes())
	}
}

func TestInnerSerialiseID(t *testing.T) {
	proto := &spiProto{
		id:      1,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	buf := new(bytes.Buffer)
	pkt := packet.Packet{
		Endpoint: 0x37,
		Data:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
	}

	buf.Reset()
	expect := []byte{0x02, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	expect = append(expect, crc8.Checksum(expect, proto.crc))
	proto.serialise(buf, pkt)
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, buf.Bytes())
	}

	buf.Reset()
	expect = []byte{0x03, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	expect = append(expect, crc8.Checksum(expect, proto.crc))
	proto.serialise(buf, pkt)
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, buf.Bytes())
	}
}

func TestInnerSerialiseMultiFrame(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	buf := new(bytes.Buffer)
	pkt := packet.Packet{
		Endpoint: 0x37,
		Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
			0x0e, 0x0f, 0x10, 0x11},
	}

	buf.Reset()
	expect := []byte{0x01, 0x37, 0x01, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	expect = append(expect, crc8.Checksum(expect, proto.crc))
	expect = append(expect, []byte{0x02, 0x37, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	expect = append(expect, crc8.Checksum(expect[9:], proto.crc))
	proto.serialise(buf, pkt)
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, buf.Bytes())
	}

	pkt = packet.Packet{
		Endpoint: 0x37,
		Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
			0x0e, 0x0f, 0x10, 0x11,
			0x12, 0x13, 0x14, 0x15},
	}
	buf.Reset()
	expect = []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	expect = append(expect, crc8.Checksum(expect, proto.crc))
	expect = append(expect, []byte{0x04, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	expect = append(expect, crc8.Checksum(expect[9:], proto.crc))
	expect = append(expect, []byte{0x05, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}...)
	expect = append(expect, crc8.Checksum(expect[18:], proto.crc))
	proto.serialise(buf, pkt)
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, buf.Bytes())
	}
}

func TestSerialise(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	pkts := []packet.Packet{
		{
			Endpoint: 0x37,
			Data:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
		},
	}
	expect := []byte{0x01, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d, 0xdd}

	res := proto.Serialise(pkts)
	if !bytes.Equal(res, expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, res)
	}
}

func TestSerialiseMultiPacket(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	pkts := []packet.Packet{
		{
			Endpoint: 0x37,
			Data:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
		},
		{
			Endpoint: 0x37,
			Data:     []byte{0x0e, 0x0f, 0x10, 0x11},
		},
	}
	expect := []byte{0x01, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	expect = append(expect, crc8.Checksum(expect, proto.crc))
	expect = append(expect, []byte{0x02, 0x37, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	expect = append(expect, crc8.Checksum(expect[9:], proto.crc))

	res := proto.Serialise(pkts)
	if !bytes.Equal(res, expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, res)
	}
}

func TestSerialiseMultiPacketMultiFrame(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	pkts := []packet.Packet{
		{
			Endpoint: 0x37,
			Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
				0x0e, 0x0f, 0x10, 0x11,
				0x12, 0x13, },
		},
		{
			Endpoint: 0x42,
			Data:     []byte{0x00, 0x01, 0x02, 0x03},
		},
	}
	expect := []byte{0x01, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	expect = append(expect, crc8.Checksum(expect, proto.crc))
	expect = append(expect, []byte{0x02, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	expect = append(expect, crc8.Checksum(expect[9:], proto.crc))
	expect = append(expect, []byte{0x03, 0x37, 0x00, 0x00, 0x12, 0x13, 0x00, 0x00}...)
	expect = append(expect, crc8.Checksum(expect[18:], proto.crc))
	expect = append(expect, []byte{0x04, 0x42, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03}...)
	expect = append(expect, crc8.Checksum(expect[27:], proto.crc))

	res := proto.Serialise(pkts)
	if !bytes.Equal(res, expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, res)
	}
}
