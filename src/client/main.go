package main

import (
	"Bitcoin/src/bitcoin/client"
	"Bitcoin/src/cryptography"
	"Bitcoin/src/protocol"
	"flag"
	"log"
	"time"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func main() {
	flag.Parse()
	req := &protocol.TransactionReq{
		InLen:  0,
		OutLen: 0,
		Ins:    []*protocol.InReq{},
		Outs:   []*protocol.OutReq{},
		Time:   time.Now().UnixMilli(),
	}
	hash, err := cryptography.Hash(req)
	if err != nil {
		log.Fatalf("hash transaction request failed: %v", err)

	}
	req.Hash = hash

	client := client.NewBitcoinClient(*addr)
	reply, err := client.SendTx(req)
	if err != nil {
		log.Fatalf("send transaction failed: %v", err)
	}
	log.Printf("send transaction result: %v", reply.Result)
}
