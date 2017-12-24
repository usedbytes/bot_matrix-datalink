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

func checkSerialise(t *testing.T, expected, data [][]byte) {
	if len(data) != len(expected) {
		t.Fatalf("Unexpected number of transfers. Expected: %d, got: %d\n",
		len(expected), len(data))
	}

	for i, _ := range(data) {
		t.Logf("Transfer %d:\n  Expected: %x\n       Got: %x\n",
			i, expected[i], data[i])

		if !bytes.Equal(expected[i], data[i]) {
			t.Fatalf("Data mismatch (%d):\n  Expected: %x\n       Got: %x\n",
				i, expected[i], data[i])
		}
	}
}

func TestSerialiseOne(t *testing.T) {
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	pkt := datalink.Packet{
		Endpoint: 0x37,
		Data:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
	}
	expected := [][]byte{
		{0x01, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d, 0xdd},
	}

	transfers := link.serialiseOne(pkt)

	checkSerialise(t, expected, transfers)
}

func addCrc(s *spiLink, d []byte) []byte {
	return append(d, crc8.Checksum(d, s.crc))
}

func TestSerialiseOneZeroPacket(t *testing.T) {
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	pkt := datalink.Packet{
		Endpoint: 0,
		Data:     nil,
	}
	expected := [][]byte{
		addCrc(link, []byte{0x01, 0, 0, 0, 0, 0, 0, 0,}),
	}

	transfers := link.serialiseOne(pkt)
	checkSerialise(t, expected, transfers)

	pkt = datalink.Packet{
		Endpoint: 0,
		Data:     []byte{},
	}
	expected = [][]byte{
		addCrc(link, []byte{0x02, 0, 0, 0, 0, 0, 0, 0,}),
	}

	transfers = link.serialiseOne(pkt)
	checkSerialise(t, expected, transfers)
}

func TestSerialiseOneShortData(t *testing.T) {
	link := &spiLink{
		id:      1,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	pkt := datalink.Packet{
		Endpoint: 0x37,
		Data:     []byte{0x0a},
	}
	expected := [][]byte{
		addCrc(link, []byte{0x02, 0x37, 0, 0, 0xa, 0, 0, 0,}),
	}

	transfers := link.serialiseOne(pkt)
	checkSerialise(t, expected, transfers)
}

func TestSerialiseOneID(t *testing.T) {
	link := &spiLink{
		id:      1,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	pkt := datalink.Packet{
		Endpoint: 0x37,
		Data:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
	}

	expected := [][]byte{
		addCrc(link, []byte{0x02, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
	}
	transfers := link.serialiseOne(pkt)
	checkSerialise(t, expected, transfers)

	expected = [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
	}
	transfers = link.serialiseOne(pkt)
	checkSerialise(t, expected, transfers)
}

func TestSerialiseOneMultiFrame(t *testing.T) {
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	pkt := datalink.Packet{
		Endpoint: 0x37,
		Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
			0x0e, 0x0f, 0x10, 0x11},
	}

	expected := [][]byte{
		addCrc(link, []byte{0x01, 0x37, 0x01, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x02, 0x37, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
	}
	transfers := link.serialiseOne(pkt)
	checkSerialise(t, expected, transfers)

	pkt = datalink.Packet{
		Endpoint: 0x37,
		Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
			0x0e, 0x0f, 0x10, 0x11,
			0x12, 0x13, 0x14, 0x15},
	}
	expected = [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x04, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x05, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}),
	}
	transfers = link.serialiseOne(pkt)
	checkSerialise(t, expected, transfers)
}

func TestSerialise(t *testing.T) {
	link := &spiLink{
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
	expected := [][]byte{
		addCrc(link, []byte{0x01, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d, }),
	}

	transfers := link.serialise(pkts)
	checkSerialise(t, expected, transfers)
}

func TestSerialiseMultiPacket(t *testing.T) {
	link := &spiLink{
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
	expected := [][]byte{
		addCrc(link, []byte{0x01, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d, }),
		addCrc(link, []byte{0x02, 0x37, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11, }),
	}

	transfers := link.serialise(pkts)
	checkSerialise(t, expected, transfers)
}

func TestSerialiseMultiPacketMultiFrame(t *testing.T) {
	link := &spiLink{
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
	expected := [][]byte{
		addCrc(link, []byte{0x01, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x02, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x03, 0x37, 0x00, 0x00, 0x12, 0x13, 0x00, 0x00}),
		addCrc(link, []byte{0x04, 0x42, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03}),
	}
	transfers := link.serialise(pkts)
	checkSerialise(t, expected, transfers)
}

func packetsEqual(a, b datalink.Packet) bool {
	if a.Endpoint != b.Endpoint {
		return false
	}

	return bytes.Equal(a.Data, b.Data)
}

func checkDeSerialise(t *testing.T, expected, data []datalink.Packet) {
	if len(data) != len(expected) {
		t.Fatalf("Unexpected number of packets. Expected: %d, got: %d\n",
		len(expected), len(data))
	}

	for i, _ := range(data) {
		t.Logf("Packet %d:\n  Expected: %x\n       Got: %x\n",
			i, expected[i], data[i])

		if !packetsEqual(expected[i], data[i]) {
			t.Fatalf("Packet mismatch (%d):\n  Expected: %x\n       Got: %x\n",
				i, expected[i], data[i])
		}
	}
}

func TestDeSerialise(t *testing.T) {
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := [][]byte{
		addCrc(link, []byte{0x19, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
	}
	expected := []datalink.Packet{
		{
			Endpoint: 0x37,
			Data:     []byte{0x0a, 0x0b, 0x0c, 0x0d},
		},
	}

	pkts, err := link.deSerialise(data)
	if err != nil {
		t.Error(err.Error())
		return
	}

	checkDeSerialise(t, expected, pkts)
}

func TestDeSerialiseBadCRC(t *testing.T) {
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := [][]byte{
		addCrc(link, []byte{0x19, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
	}
	data[0][0] += 1

	pkts, err := link.deSerialise(data)
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
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	expected := []datalink.Packet{
		{
			Endpoint: 0x37,
			Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
			0x0e, 0x0f, 0x10, 0x11,
			0x12, 0x13, 0x14, 0x15},
		},
	}

	data := [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x04, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x05, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}),
	}

	pkts, err := link.deSerialise(data)
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkDeSerialise(t, expected, pkts)
}

func TestDeSerialiseMultiPacket(t *testing.T) {
	link := &spiLink{
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
	data := [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x01, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x04, 0x37, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x05, 0x38, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}),
	}

	pkts, err := link.deSerialise(data)
	if err != nil {
		t.Error(err.Error())
		return
	}
	checkDeSerialise(t, expected, pkts)
}

func TestDeSerialiseBadID(t *testing.T) {
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x00, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x08, 0x38, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x04, 0x39, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}),
	}

	_, err := link.deSerialise(data)
	if err == nil {
		t.Errorf("(Multi packet) Expected error, got none.\n")
	} else if !strings.HasPrefix(err.Error(), "Invalid packet ID") {
		t.Errorf("(Multi packet) Unexpected error, expected 'Invalid packet ID', got: %s.\n",
			 err.Error())
	}

	data = [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x08, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x04, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}),
	}

	_, err = link.deSerialise(data)
	if err == nil {
		t.Errorf("(Single packet) Expected error, got none.\n")
	} else if !strings.HasPrefix(err.Error(), "Invalid packet ID") {
		t.Errorf("(Single packet) Unexpected error, expected 'Invalid packet ID', got: %s.\n",
			 err.Error())
	}
}

func TestDeSerialiseBadEndpoint(t *testing.T) {
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x04, 0x38, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x04, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}),
	}

	_, err := link.deSerialise(data)
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
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x04, 0x37, 0x00, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x04, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}),
	}

	_, err := link.deSerialise(data)
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
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	data := [][]byte{
		{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, },
	}
	_, err := link.deSerialise(data)
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
	link := &spiLink{
		id:      0,
		datalen: 4,
		crc:     crc8.MakeTable(crc8.CRC8),
	}

	expected := []datalink.Packet{
		{
			Endpoint: 0x37,
			Data: []byte{0x0a, 0x0b, 0x0c, 0x0d,
			0x0e, 0x0f, 0x10, 0x11,
			0x12, 0x13, 0x14, 0x15},
		},
	}
	data := [][]byte{
		addCrc(link, []byte{0x03, 0x37, 0x02, 0x00, 0x0a, 0x0b, 0x0c, 0x0d}),
		addCrc(link, []byte{0x04, 0x37, 0x01, 0x00, 0x0e, 0x0f, 0x10, 0x11}),
		addCrc(link, []byte{0x05, 0x37, 0x00, 0x00, 0x12, 0x13, 0x14, 0x15}),
	}

	pkts, err := link.deSerialise(data[:1])
	if err != nil {
		t.Error(err.Error())
		return
	}

	if len(pkts) != 0 {
		t.Errorf("Unexpected number of packets. Expected: %d, got: %d\n",
			 0, len(pkts))
		return
	}

	pkts, err = link.deSerialise(data[1:])
	if err != nil {
		t.Error(err.Error())
		return
	}

	checkDeSerialise(t, expected, pkts)
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
