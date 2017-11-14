// Copyright 2017 Brian Starkey <stark3y@gmail.com>
package rpcconn

import (
	"net"
	"net/rpc"
	"github.com/usedbytes/bot_matrix/datalink"
)

type RPCEndpoint struct {
	transactor datalink.Transactor
}

type RPCServ struct {
	endpoint RPCEndpoint

	srv *rpc.Server
}

func (r *RPCEndpoint) RPCTransact(tx []datalink.Packet, rx *[]datalink.Packet) error {
	pkts, err := r.transactor.Transact(tx)

	*rx = pkts

	return err
}

func (r *RPCServ) Serve(l net.Listener) {
	r.srv.Accept(l)
}

func NewRPCServ(conn datalink.Transactor) (*RPCServ, error) {
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

func (c *RPCClient) Transact(tx []datalink.Packet) ([]datalink.Packet, error) {
	rx := make([]datalink.Packet, 0, len(tx))
	err := c.client.Call("RPCEndpoint.RPCTransact", tx, &rx)

	return rx, err
}
