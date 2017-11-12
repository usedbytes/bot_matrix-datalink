package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/usedbytes/bot_matrix/datalink/connection"
	"github.com/usedbytes/bot_matrix/datalink/packet"
	"github.com/usedbytes/bot_matrix/datalink/spiconn"
)

func ledOn(c *connection.Connection) {
	data := []packet.Packet{
		{ 1, []byte{1} },
	}
	c.Transact(data)
}

func ledOff(c *connection.Connection) {
	data := []packet.Packet{
		{ 1, []byte{0} },
	}
	c.Transact(data)
}

func setFreq(c *connection.Connection, freq uint32) {
	data := []packet.Packet{
		{ Endpoint: 2, },
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, freq)

	data[0].Data = buf.Bytes()

	c.Transact(data)
}

func setDuty(c *connection.Connection, ch byte, dir byte, duty uint16) {
	data := []packet.Packet{
		{ Endpoint: 3, },
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, ch)
	binary.Write(buf, binary.LittleEndian, dir)
	binary.Write(buf, binary.LittleEndian, duty)

	data[0].Data = buf.Bytes()

	c.Transact(data)
}

func setGains(c *connection.Connection, Kc, Kd, Ki float64) {
	data := []packet.Packet{
		{ Endpoint: 4, },
	}

	iKc := int32(Kc * 65536)
	iKd := int32(Kd * 65536)
	iKi := int32(Ki * 65536)

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, iKc)
	binary.Write(buf, binary.LittleEndian, iKd)
	binary.Write(buf, binary.LittleEndian, iKi)

	data[0].Data = buf.Bytes()

	c.Transact(data)
}

func setPoint(c *connection.Connection, sp uint32) {
	data := []packet.Packet{
		{ Endpoint: 5, },
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, sp)

	data[0].Data = buf.Bytes()

	c.Transact(data)
}

func setIlimit(c *connection.Connection, il uint32) {
	data := []packet.Packet{
		{ Endpoint: 6, },
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, il)

	data[0].Data = buf.Bytes()

	c.Transact(data)
}

func main() {
	var on bool
	c, _ := spiconn.NewSPIConn("/dev/spidev0.0")

	// create new shell.
	// by default, new shell includes 'exit', 'help' and 'clear' commands.
	shell := ishell.New()

	// display welcome info.
	shell.Println("Driver...")

	// register a function for "greet" command.
	shell.AddCmd(&ishell.Cmd{
		Name: "sp",
		Help: "set_point",
		Func: func(ctx *ishell.Context) {
			if len(ctx.Args) != 1 {
				ctx.Err(fmt.Errorf("Expected one argument (setpoint uint32)"))
				return
			}

			sp, err := strconv.ParseUint(ctx.Args[0], 0, 32)
			if err != nil {
				ctx.Err(err)
				return
			}

			setPoint(c, uint32(sp))
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "il",
		Help: "ilimit",
		Func: func(ctx *ishell.Context) {
			if len(ctx.Args) != 1 {
				ctx.Err(fmt.Errorf("Expected one argument (ilimi int32)"))
				return
			}

			il, err := strconv.ParseUint(ctx.Args[0], 0, 32)
			if err != nil {
				ctx.Err(err)
				return
			}

			setIlimit(c, uint32(il))
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "g",
		Help: "gains",
		Func: func(ctx *ishell.Context) {
			if len(ctx.Args) != 3 {
				ctx.Err(fmt.Errorf("Expected three arguments (Kc, Kd, Ki)"))
				return
			}

			Kc, err := strconv.ParseFloat(ctx.Args[0], 64)
			if err != nil {
				ctx.Err(err)
				return
			}

			Kd, err := strconv.ParseFloat(ctx.Args[1], 64)
			if err != nil {
				ctx.Err(err)
				return
			}

			Ki, err := strconv.ParseFloat(ctx.Args[2], 64)
			if err != nil {
				ctx.Err(err)
				return
			}

			setGains(c, Kc, Kd, Ki)
		},
	})

	// run shell
	shell.Run()

	for {
		for f := uint32(2000); f < 20000; f += 1000 {
			setFreq(c, f);
			for rev := byte(0); rev <= 1; rev++ {
				fmt.Printf("rev: %d\n", rev)
				for d := uint16(0); d < 65535 - 300; d+=300 {
					setDuty(c, 0, rev, d)
					time.Sleep(30 * time.Millisecond)
				}
			}

			if on {
				ledOn(c)
			} else {
				ledOff(c)
			}
			on = !on
		}
	}
}
