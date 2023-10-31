package main

import (
	"Bitcoin/src/bitcoin/server"
	"Bitcoin/src/config"
	"Bitcoin/src/protocol"
	"flag"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	CONFIG = flag.String("config", "config.yml", "the path of config file")
)

func main() {
	flag.Parse()

	cfg, err := config.Read(*CONFIG)
	if err != nil {
		log.Fatalf("read config error: %v", err)
	}

	listener, err := net.Listen("tcp", cfg.Endpoint)
	if err != nil {
		log.Fatalf("failed to listen %v: %v", cfg.Endpoint, err)
	}

	register := grpc.NewServer()
	server, err := server.NewBitcoinServer(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	go server.MineBlock()
	go server.BroadcastTx()
	go server.BroadcastBlock()

	protocol.RegisterTransactionServer(register, server)
	log.Printf("server listening at %v", listener.Addr())

	if err := register.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
