// Copyright 2017 Brian Starkey <stark3y@gmail.com>

package spiconn

import (
	"bytes"
	"flag"
	"strings"
	"testing"

	"github.com/sigurn/crc8"
	"github.com/usedbytes/bot_matrix/datalink"
)

func TestInnerSerialise(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	buf := new(bytes.Buffer)
	pkt := datalink.Packet{
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

func TestInnerSerialiseZeroPacket(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	buf := new(bytes.Buffer)
	pkt := datalink.Packet{
		Endpoint: 0,
		Data:     nil,
	}
	expect := []byte{0x01, 0, 0, 0, 0, 0, 0, 0,}
	expect = append(expect, crc8.Checksum(expect, proto.crc))

	proto.serialise(buf, pkt)
	if !bytes.Equal(buf.Bytes(), expect) {
		t.Errorf("Data mismatch:\n  Expected: %x\n       Got: %x\n",
			expect, buf.Bytes())
	}

	buf = new(bytes.Buffer)
	pkt = datalink.Packet{
		Endpoint: 0,
		Data:     []byte{},
	}
	expect = []byte{0x02, 0, 0, 0, 0, 0, 0, 0,}
	expect = append(expect, crc8.Checksum(expect, proto.crc))

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
	pkt := datalink.Packet{
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
	pkt := datalink.Packet{
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
	pkt := datalink.Packet{
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

	pkt = datalink.Packet{
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

	pkts := []datalink.Packet{
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

	pkts := []datalink.Packet{
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

	pkts := []datalink.Packet{
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

func packetsEqual(a, b datalink.Packet) bool {
	if a.Endpoint != b.Endpoint {
		return false
	}

	return bytes.Equal(a.Data, b.Data)
}

func TestDeSerialise(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := []byte{0x19, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc))
	expected := datalink.Packet{
		Endpoint: 0x37,
		Data:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
	}

	pkts, err := proto.DeSerialise(data)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(pkts) != 1 {
		t.Errorf("Unexpected number of packets. Expected: %d, got: %d\n",
			 1, len(pkts))
		return
	}

	if !packetsEqual(expected, pkts[0]) {
		t.Errorf("Packet mismatch.\n  Expected: %v\n       Got: %v\n",
			 expected, pkts[0])
		return

	}
}

func TestDeSerialiseBadCRC(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := []byte{0x19, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc) - 1)

	pkts, err := proto.DeSerialise(data)
	if err == nil {
		t.Errorf("Expected error, got none.\n")
		return
	}

	if !strings.HasPrefix(err.Error(), "CRC error") {
		t.Errorf("Unexpected error, expected 'CRC error', got: %s.\n",
			 err.Error())
		return
	}

	if len(pkts) != 0 {
		t.Errorf("Unexpected number of packets. Expected: %d, got: %d\n",
			 0, len(pkts))
		return
	}
}

func TestDeSerialiseMultiFrame(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	expected := datalink.Packet{
		Endpoint: 0x37,
		Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
		0x0e, 0x0f, 0x10, 0x11,
		0x12, 0x13, 0x14, 0x15},
	}
	data := []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc))
	data = append(data, []byte{0x04, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	data = append(data, crc8.Checksum(data[9:], proto.crc))
	data = append(data, []byte{0x05, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}...)
	data = append(data, crc8.Checksum(data[18:], proto.crc))

	pkts, err := proto.DeSerialise(data)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(pkts) != 1 {
		t.Errorf("Unexpected number of packets. Expected: %d, got: %d\n",
			 1, len(pkts))
		return
	}

	if !packetsEqual(expected, pkts[0]) {
		t.Errorf("Packet mismatch.\n  Expected: %v\n       Got: %v\n",
			 expected, pkts[0])
		return

	}
}

func TestDeSerialiseMultiPacket(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	expected := []datalink.Packet{
		{
			Endpoint: 0x37,
			Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
			0x0e, 0x0f, 0x10, 0x11},
		},
		{
			Endpoint: 0x38,
			Data: []byte{0x12, 0x13, 0x14, 0x15},
		},
	}
	data := []byte{0x03, 0x37, 0x01, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc))
	data = append(data, []byte{0x04, 0x37, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	data = append(data, crc8.Checksum(data[9:], proto.crc))
	data = append(data, []byte{0x05, 0x38, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}...)
	data = append(data, crc8.Checksum(data[18:], proto.crc))

	pkts, err := proto.DeSerialise(data)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(pkts) != 2 {
		t.Errorf("Unexpected number of packets. Expected: %d, got: %d\n",
			 2, len(pkts))
		return
	}

	if !packetsEqual(expected[0], pkts[0]) {
		t.Errorf("Packet mismatch.\n  Expected: %v\n       Got: %v\n",
			 expected[0], pkts[0])
		return
	}

	if !packetsEqual(expected[1], pkts[1]) {
		t.Errorf("Packet mismatch.\n  Expected: %v\n       Got: %v\n",
			 expected[1], pkts[1])
		return
	}
}

func TestDeSerialiseBadID(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := []byte{0x03, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc))
	data = append(data, []byte{0x08, 0x38, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	data = append(data, crc8.Checksum(data[9:], proto.crc))
	data = append(data, []byte{0x04, 0x39, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}...)
	data = append(data, crc8.Checksum(data[18:], proto.crc))

	_, err := proto.DeSerialise(data)
	if err == nil {
		t.Errorf("(Multi packet) Expected error, got none.\n")
	} else if !strings.HasPrefix(err.Error(), "Invalid packet ID") {
		t.Errorf("(Multi packet) Unexpected error, expected 'Invalid packet ID', got: %s.\n",
			 err.Error())
	}

	data = []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc))
	data = append(data, []byte{0x08, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	data = append(data, crc8.Checksum(data[9:], proto.crc))
	data = append(data, []byte{0x04, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}...)
	data = append(data, crc8.Checksum(data[18:], proto.crc))

	_, err = proto.DeSerialise(data)
	if err == nil {
		t.Errorf("(Single packet) Expected error, got none.\n")
	} else if !strings.HasPrefix(err.Error(), "Invalid packet ID") {
		t.Errorf("(Single packet) Unexpected error, expected 'Invalid packet ID', got: %s.\n",
			 err.Error())
	}
}

func TestDeSerialiseBadEndpoint(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc))
	data = append(data, []byte{0x04, 0x38, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	data = append(data, crc8.Checksum(data[9:], proto.crc))
	data = append(data, []byte{0x05, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}...)
	data = append(data, crc8.Checksum(data[18:], proto.crc))

	_, err := proto.DeSerialise(data)
	if err == nil {
		t.Errorf("Expected error, got none.\n")
		return
	}

	if !strings.HasPrefix(err.Error(), "Invalid Endpoint") {
		t.Errorf("Unexpected error, expected 'Invalid Endpoint', got: %s.\n",
			 err.Error())
		return
	}
}

func TestDeSerialiseBadNparts(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc))
	data = append(data, []byte{0x04, 0x37, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	data = append(data, crc8.Checksum(data[9:], proto.crc))
	data = append(data, []byte{0x05, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}...)
	data = append(data, crc8.Checksum(data[18:], proto.crc))

	_, err := proto.DeSerialise(data)
	if err == nil {
		t.Errorf("Expected error, got none.\n")
		return
	}

	if !strings.HasPrefix(err.Error(), "Invalid nparts") {
		t.Errorf("Unexpected error, expected 'Invalid nparts', got: %s.\n",
			 err.Error())
		return
	}
}

func TestDeSerialiseShortData(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, }
	_, err := proto.DeSerialise(data)
	if err == nil {
		t.Errorf("Expected error, got none.\n")
		return
	}

	if !strings.HasPrefix(err.Error(), "Short data") {
		t.Errorf("Unexpected error, expected 'Short data', got: %s.\n",
			 err.Error())
		return
	}
}

func TestDeSerialiseSplitFrames(t *testing.T) {
	proto := &spiProto{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	expected := datalink.Packet{
		Endpoint: 0x37,
		Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
		0x0e, 0x0f, 0x10, 0x11,
		0x12, 0x13, 0x14, 0x15},
	}
	data := []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}
	data = append(data, crc8.Checksum(data, proto.crc))
	data = append(data, []byte{0x04, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}...)
	data = append(data, crc8.Checksum(data[9:], proto.crc))
	data = append(data, []byte{0x05, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}...)
	data = append(data, crc8.Checksum(data[18:], proto.crc))

	pkts, err := proto.DeSerialise(data[:9])
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(pkts) != 0 {
		t.Errorf("Unexpected number of packets. Expected: %d, got: %d\n",
			 0, len(pkts))
		return
	}

	pkts, err = proto.DeSerialise(data[9:])
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(pkts) != 1 {
		t.Errorf("Unexpected number of packets. Expected: %d, got: %d\n",
			 1, len(pkts))
		return
	}

	if !packetsEqual(expected, pkts[0]) {
		t.Errorf("Packet mismatch.\n  Expected: %v\n       Got: %v\n",
			 expected, pkts[0])
		return
	}
}

var devname string

func init() {
	flag.StringVar(&devname, "devname", "", "spidev device to use for loopback test")
	flag.Parse()
}

func TestSpiconnConnection(t *testing.T) {
	if len(devname) == 0 {
		t.SkipNow()
	}

	conn, err := NewSPIConn(devname)
	if err != nil {
		t.Error(err)
		return
	}

	pkts := []datalink.Packet{
		{
			Endpoint: 0x37,
			Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
				0x0e, 0x0f, 0x10, 0x11,
				0x12, 0x13, 0x00, 0x00},
		},
		{
			Endpoint: 0x42,
			Data:     []byte{0x00, 0x01, 0x02, 0x03},
		},
	}
	res, err := conn.Transact(pkts)
	if err != nil {
		t.Error(err)
		return
	}

	if len(res) != len(pkts) {
		t.Errorf("Unexpected number of packets. Expected %d got %d\n",
			len(pkts), len(res))
		return
	}

	for i, p := range res {
		if !packetsEqual(pkts[i], p) {
			t.Errorf("Packet %d mismatch.\n  Expected: %v\n       Got: %v\n",
				 i, pkts[i], p)
			return
		}

	}
}
