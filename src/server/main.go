package main

import (
	"Bitcoin/src/config"
	"Bitcoin/src/protocol"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	PORT   = flag.Int("port", 50051, "The server port")
	CONFIG = "config.yml"
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *PORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	cfg, err := config.Read(CONFIG)
	if err != nil {
		log.Fatalf("read config error: %v", err)
	}

	register := grpc.NewServer()
	server, err := NewBitcoinServer(cfg.DB)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	go server.broadcastTx()

	protocol.RegisterTransactionServer(register, server)
	log.Printf("server listening at %v", listener.Addr())

	if err := register.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
