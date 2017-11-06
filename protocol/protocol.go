// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package protocol

import (
	"github.com/usedbytes/bot_matrix/datalink/packet"
)

type Protocol interface {
	Serialise([]packet.Packet) []byte
	DeSerialise([]byte) ([]packet.Packet, error)
}
