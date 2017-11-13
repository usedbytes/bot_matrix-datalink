// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package rpcconn

import (
	"net"
	"net/rpc"
	"github.com/usedbytes/bot_matrix/datalink/connection"
	"github.com/usedbytes/bot_matrix/datalink/packet"
)

type RPCEndpoint struct {
	transactor connection.Transactor
}

type RPCServ struct {
	endpoint RPCEndpoint

	srv *rpc.Server
}

func (r *RPCEndpoint) RPCTransact(tx []packet.Packet, rx *[]packet.Packet) error {
	pkts, err := r.transactor.Transact(tx)

	*rx = pkts

	return err
}

func (r *RPCServ) Serve(l net.Listener) {
	r.srv.Accept(l)
}

func NewRPCServ(conn connection.Transactor) (*RPCServ, error) {
	srv := &RPCServ{ endpoint: RPCEndpoint{ conn } }

	srv.srv = rpc.NewServer()
	srv.srv.Register(&srv.endpoint)

	return srv, nil
}

type RPCClient struct {
	client *rpc.Client
}

func NewRPCClient(server string) (*RPCClient, error) {
	client, err := rpc.Dial("tcp", server)
	if err != nil {
		return nil, err
	}

	return &RPCClient{ client }, nil
}

func (c *RPCClient) Transact(tx []packet.Packet) ([]packet.Packet, error) {
	rx := make([]packet.Packet, 0, len(tx))
	err := c.client.Call("RPCEndpoint.RPCTransact", tx, &rx)

	return rx, err
}
