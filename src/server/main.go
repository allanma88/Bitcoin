package main

import (
	"Bitcoin/src/bitcoin/server"
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/protocol"
	"context"
	"flag"
	"log"
	"net"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
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

	db, err := leveldb.OpenFile(cfg.DataDir, nil)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	blockdb := database.NewBlockDB(db)
	blockContentDb := database.NewBlockContentDB(db)

	ctx, cancelFunc := context.WithCancelCause(context.Background())

	server, err := server.NewBitcoinServer(cfg, blockdb, blockContentDb, cancelFunc)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	wg := &sync.WaitGroup{}

	go server.MineBlock(ctx, wg)
	go server.BroadcastTx()
	go server.BroadcastBlock()
	go server.SyncBlocks(wg)

	listener, err := net.Listen("tcp", cfg.Endpoint)
	if err != nil {
		log.Fatalf("failed to listen %v: %v", cfg.Endpoint, err)
	}

	register := grpc.NewServer()
	protocol.RegisterTransactionServer(register, server)
	protocol.RegisterBlockServer(register, server)
	log.Printf("server listening at %v", listener.Addr())

	if err := register.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
