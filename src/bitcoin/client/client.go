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

func (service *BitcoinClient) SendTx(req *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	if service.conn == nil || service.conn.GetState() == connectivity.Shutdown {
		// Set up a connection to the server.
		conn, err := grpc.Dial(service.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		log.Printf("connected to %s", service.addr)
		service.conn = conn
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := protocol.NewTransactionClient(service.conn)
	return client.ExecuteTx(ctx, req)
}
