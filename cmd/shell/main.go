package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/abiosoft/ishell"
	"github.com/usedbytes/bot_matrix/datalink"
	"github.com/usedbytes/bot_matrix/datalink/spiconn"
	"github.com/usedbytes/bot_matrix/datalink/rpcconn"
)

type motor_set struct {
	Direction int32
	SetPoint uint32
}

type motor_cmd_set struct {
	A motor_set
	B motor_set
}

func pumpDatalink(conn datalink.Transactor, tx <-chan datalink.Packet,
		  rx chan<- datalink.Packet, stop <-chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)

	minNum := 4
	toSend := make([]datalink.Packet, 0, minNum)

	for {
		select {
		case _ = <-ticker.C:
			if len(toSend) > 0 {
				fmt.Printf("Have %d packets to send.\n", len(toSend))
			}
			if len(toSend) < minNum {
				toSend = append(toSend, make([]datalink.Packet, minNum - len(toSend))...)
			}

			pkts, err := conn.Transact(toSend)
			if err != nil {
				fmt.Printf("Error! %s\n", err)
				time.Sleep(500 * time.Millisecond)
			} else {
				for _, p := range pkts {
					rx <-p
				}
			}

			toSend = make([]datalink.Packet, 0, minNum)

		case p := <-tx:
			toSend = append(toSend, p)

		case _ = <-stop:
			return
		}
	}
}

func ledOn(c datalink.Transactor) {
	data := []datalink.Packet{
		{ 1, []byte{1} },
	}
	c.Transact(data)
}

func ledOff(c datalink.Transactor) {
	data := []datalink.Packet{
		{ 1, []byte{0} },
	}
	c.Transact(data)
}

func reset(c datalink.Transactor) {
	data := []datalink.Packet{
		{ 0xfe, []byte{1} },
	}
	c.Transact(data)
}

func setFreq(c datalink.Transactor, freq uint32) {
	data := []datalink.Packet{
		{ Endpoint: 2, },
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, freq)

	data[0].Data = buf.Bytes()

	c.Transact(data)
}

func setDuty(c datalink.Transactor, ch byte, dir byte, duty uint16) {
	data := []datalink.Packet{
		{ Endpoint: 3, },
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, ch)
	binary.Write(buf, binary.LittleEndian, dir)
	binary.Write(buf, binary.LittleEndian, duty)

	data[0].Data = buf.Bytes()

	c.Transact(data)
}

func setGains(c datalink.Transactor, Kc, Kd, Ki float64) {
	data := []datalink.Packet{
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

func setPoint(c datalink.Transactor, sp uint32) {
	set := motor_cmd_set{
		A: motor_set{ Direction: 0, SetPoint: sp },
		B: motor_set{ Direction: 1, SetPoint: sp },
	}

	data := []datalink.Packet{
		{ Endpoint: 18, },
	}


	data[0].Data, _ = set.MarshalBinary()

	c.Transact(data)
}

func setIlimit(c datalink.Transactor, il uint32) {
	data := []datalink.Packet{
		{ Endpoint: 6, },
	}

	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, il)

	data[0].Data = buf.Bytes()

	c.Transact(data)
}

func (m *motor_cmd_set) MarshalBinary() (data []byte, err error) {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, int32(0))
	binary.Write(buf, binary.LittleEndian, m.A.Direction)
	binary.Write(buf, binary.LittleEndian, m.A.SetPoint)
	binary.Write(buf, binary.LittleEndian, m.B.Direction)
	binary.Write(buf, binary.LittleEndian, m.B.SetPoint)

	return buf.Bytes(), nil
}

type mdata struct {
	Count uint32
	SetPoint uint32
	Duty uint16
	Enabling uint16
}

type motor_data struct {
	Timestamp uint32
	A mdata
	B mdata
}

func (m *motor_data) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &m.Timestamp)
	binary.Read(buf, binary.LittleEndian, &m.A.Count)
	binary.Read(buf, binary.LittleEndian, &m.A.SetPoint)
	binary.Read(buf, binary.LittleEndian, &m.A.Duty)
	binary.Read(buf, binary.LittleEndian, &m.A.Enabling)
	binary.Read(buf, binary.LittleEndian, &m.B.Count)
	binary.Read(buf, binary.LittleEndian, &m.B.SetPoint)
	binary.Read(buf, binary.LittleEndian, &m.B.Duty)
	binary.Read(buf, binary.LittleEndian, &m.B.Enabling)

	return nil
}

func characteriseMotor(c datalink.Transactor, ctx *ishell.Context) {
	tx := make(chan datalink.Packet, 10)
	rx := make(chan datalink.Packet, 10)
	stop := make(chan bool, 1)

	f, err := os.Create("log.csv")
	if err != nil {
		ctx.Err(err)
		return
	}
	defer f.Close()

	ticker := time.NewTicker(time.Second * 2)

	go pumpDatalink(c, tx, rx, stop)

	for duty := uint16(0); duty <= 65535 - 1000; {
		select {
		case _ = <-ticker.C:
			duty += 1000
			data := []datalink.Packet{
				{ Endpoint: 3, },
			}

			buf := &bytes.Buffer{}
			binary.Write(buf, binary.LittleEndian, byte(0))
			binary.Write(buf, binary.LittleEndian, byte(0))
			binary.Write(buf, binary.LittleEndian, duty)

			data[0].Data = buf.Bytes()
			tx <- data[0]

			fmt.Printf("Waiting for duty to change\n")
			for p := range rx {
				if p.Endpoint != 0xf {
					continue
				}

				var m motor_data;

				m.UnmarshalBinary(p.Data)
				if m.A.Duty == duty {
					break
				}
			}
			fmt.Printf("Changed to %d\n", duty)
		case p := <-rx:
			if p.Endpoint != 0xf {
				break
			}

			var m motor_data;
			m.UnmarshalBinary(p.Data)
			fmt.Fprintf(f, "%v, %v, %v\n", m.Timestamp, m.A.Count, m.A.Duty);
			fmt.Printf("%v, %v, %v\n", m.Timestamp, m.A.Count, m.A.Duty);
		}
	}

	setDuty(c, 0, 0, 0)

	stop <- true
}

const html = `
<!DOCTYPE HTML>
<html>
<head>
<script>
window.onload = function () {

var dps = []; // dataPoints
var dps2 = []; // dataPoints
var dps3 = []; // dataPoints
var dps4 = []; // dataPoints
var dps5 = []; // dataPoints
var dps6 = []; // dataPoints
var dps7 = []; // dataPoints
var dps8 = []; // dataPoints
var chart = new CanvasJS.Chart("chartContainer", {
	title :{
		text: "Dynamic Data"
	},
	axisY:[{
		title: "Actual speed (1/speed)",
		lineColor: "blue",
		titleFontColor: "blue",
		labelFontColor: "blue",
		minimum: 30,
		maximum: 1300,
	},
	{
		title: "PWM Output",
		lineColor: "red",
		titleFontColor: "red",
		labelFontColor: "red"
	},
	{
		title: "PID Setpoint",
		lineColor: "green",
		titleFontColor: "green",
		labelFontColor: "green",
		minimum: 30,
		maximum: 1300,
	},
	{
		title: "Enabling",
		lineColor: "black",
		titleFontColor: "black",
		labelFontColor: "black",
		minimum: 0,
		maximum: 20,
	},
	],
	data: [{
		axisYIndex: 0,
		color: "#cc0000",
		type: "line",
		dataPoints: dps
	},
	{
		axisYIndex: 1,
		color: "#ff8080",
		type: "line",
		dataPoints: dps2
	},
	{
		axisYIndex: 2,
		color: "#800000",
		type: "line",
		dataPoints: dps3
	},
	{
		axisYIndex: 0,
		color: "#0000cc",
		type: "line",
		dataPoints: dps4
	},
	{
		axisYIndex: 1,
		color: "#8080ff",
		type: "line",
		dataPoints: dps5
	},
	{
		axisYIndex: 2,
		color: "#000080",
		type: "line",
		dataPoints: dps6
	},
	{
		axisYIndex: 3,
		color: "#80ff80",
		type: "line",
		dataPoints: dps7
	},
	{
		axisYIndex: 3,
		color: "#00ff00",
		type: "line",
		dataPoints: dps8
	}
	]
});

var xVal = 0;
var yVal = 100; 
var updateInterval = 1000;
var dataLength = 200; // number of dataPoints visible at any point

var updateChart = function (count) {

	count = count || 1;

	for (var j = 0; j < count; j++) {
		yVal = yVal +  Math.round(5 + Math.random() *(-5-5));
		dps.push({
			x: xVal,
			y: yVal
		});
		xVal++;
	}

	if (dps.length > dataLength) {
		dps.shift();
	}

	chart.render();
};

//updateChart(dataLength);
//setInterval(function(){updateChart()}, updateInterval);

var exampleSocket = new WebSocket("ws://localhost:8080/ws");
exampleSocket.onmessage = function (event) {
	var obj = JSON.parse(event.data)

	d = {
		x: Number(obj.Timestamp),
		y: Number(obj.A.Count),
	}
	dps.push(d);
	if (dps.length > dataLength) {
		dps.shift();
	}

	d = {
		x: Number(obj.Timestamp),
		y: Number(obj.A.Duty),
	}
	dps2.push(d);
	if (dps2.length > dataLength) {
		dps2.shift();
	}

	d = {
		x: Number(obj.Timestamp),
		y: Number(obj.A.SetPoint),
	}
	dps3.push(d);
	if (dps3.length > dataLength) {
		dps3.shift();
	}


	d = {
		x: Number(obj.Timestamp),
		y: Number(obj.B.Count),
	}
	dps4.push(d);
	if (dps4.length > dataLength) {
		dps4.shift();
	}

	d = {
		x: Number(obj.Timestamp),
		y: Number(obj.B.Duty),
	}
	dps5.push(d);
	if (dps5.length > dataLength) {
		dps5.shift();
	}

	d = {
		x: Number(obj.Timestamp),
		y: Number(obj.B.SetPoint),
	}
	dps6.push(d);
	if (dps6.length > dataLength) {
		dps6.shift();
	}

	d = {
		x: Number(obj.Timestamp),
		y: Number(obj.A.Enabling),
	}
	dps7.push(d);
	if (dps7.length > dataLength) {
		dps7.shift();
	}

	d = {
		x: Number(obj.Timestamp),
		y: Number(obj.B.Enabling),
	}
	dps8.push(d);
	if (dps8.length > dataLength) {
		dps8.shift();
	}

	chart.render();
}

}

</script>
</head>
<body>
<div id="chartContainer" style="height: 370px; width:100%;"></div>
<script src="https://canvasjs.com/assets/script/canvasjs.min.js"></script>
</body>
</html>
`

func socket(ws *websocket.Conn) {
	for m := range telem {
		websocket.JSON.Send(ws, m)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, html)
}

func webserv() {
    http.HandleFunc("/", handler)
    http.Handle("/ws", websocket.Handler(socket))
    http.ListenAndServe(":8080", nil)
}

var telem chan motor_data

type wrapper struct {
	ch chan datalink.Packet
}

func (w *wrapper) Transact(tx []datalink.Packet) ([]datalink.Packet, error) {
	for _, p := range tx {
		w.ch <- p
	}

	return nil, nil
}

func main() {
	var on bool
	var devname string
	var c datalink.Transactor
	var err error

	flag.StringVar(&devname, "devname", "/dev/spidev0.0", "Device to use for communication. Use tcp:.... for RPCConn")
	flag.Parse()

	if strings.HasPrefix(devname, "tcp:") {
		c, err = rpcconn.NewRPCClient(devname[len("tcp:"):])
	} else {
		c, err = spiconn.NewSPIConn(devname)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	telem = make(chan motor_data, 10);
	go webserv();

	tx := make(chan datalink.Packet, 10)
	rx := make(chan datalink.Packet, 10)
	stop := make(chan bool, 1)

	go func() {
		for p := range rx {
			if p.Endpoint != 0xf {
				continue
			}

			var m motor_data;

			m.UnmarshalBinary(p.Data)

			select {
			case telem <- m:
			default:
				fmt.Println("drop.")
			}
		}
	}()

	go pumpDatalink(c, tx, rx, stop)
	c = &wrapper{ tx }

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

	/*
	shell.AddCmd(&ishell.Cmd{
		Name: "pump",
		Help: "pump",
		Func: func(ctx *ishell.Context) {

			close(rx)
			close(tx)
			close(stop)
		},
	})
	*/

	shell.AddCmd(&ishell.Cmd{
		Name: "d",
		Help: "duty",
		Func: func(ctx *ishell.Context) {
			if len(ctx.Args) != 1 {
				ctx.Err(fmt.Errorf("Expected one argument (setpoint uint32)"))
				return
			}

			sp, err := strconv.ParseUint(ctx.Args[0], 0, 16)
			if err != nil {
				ctx.Err(err)
				return
			}

			setDuty(c, 0, 0, uint16(sp))
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

	shell.AddCmd(&ishell.Cmd{
		Name: "char",
		Help: "char",
		Func: func(ctx *ishell.Context) {
			characteriseMotor(c, ctx)
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "reset",
		Help: "reset",
		Func: func(ctx *ishell.Context) {
			reset(c)
		},
	})

	// run shell
	shell.Run()

	return

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
