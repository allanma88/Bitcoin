package client

import (
	"Bitcoin/src/protocol"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	SENDTIMEOUT = 10 * time.Second
)

type IBitcoinClient interface {
	SendTx(req *protocol.TransactionReq) (*protocol.TransactionReply, error)
}

type BitcoinClient struct {
	addr string
	conn *grpc.ClientConn
}

func NewBitcoinClient(addr string) IBitcoinClient {
	return &BitcoinClient{addr: addr}
}

func (cli *BitcoinClient) SendTx(req *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	if cli.conn == nil || cli.conn.GetState() == connectivity.Shutdown {
		// Set up a connection to the server.
		conn, err := grpc.Dial(cli.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		log.Printf("connected to %s", cli.addr)
		cli.conn = conn
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), SENDTIMEOUT)
	defer cancel()

	client := protocol.NewTransactionClient(cli.conn)
	return client.ExecuteTx(ctx, req)
}
