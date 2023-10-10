package main

import (
	"Bitcoin/src/cryptography"
	"Bitcoin/src/protocol"
	"context"
	"flag"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect %s failed: %v", *addr, err)
	}
	log.Printf("connected to %s", *addr)
	defer conn.Close()
	client := protocol.NewTransactionClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &protocol.TransactionReq{
		InLen:     0,
		OutLen:    0,
		Ins:       []*protocol.InReq{},
		Outs:      []*protocol.OutReq{},
		Timestamp: timestamppb.Now(),
	}
	hash, err := cryptography.Hash(req)
	if err != nil {
		log.Fatalf("hash transaction request failed: %v", err)

	}
	req.Id = hash

	reply, err := client.ExecuteTx(ctx, req)
	if err != nil {
		log.Fatalf("send transaction failed: %v", err)
	}
	log.Printf("send transaction result: %v", reply.Result)
}
