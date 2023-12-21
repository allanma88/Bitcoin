package main

import (
	"Bitcoin/src/bitcoin/server"
	"Bitcoin/src/config"
	"Bitcoin/src/database"
	"Bitcoin/src/protocol"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	server, err := server.NewBitcoinServer(cfg, blockdb)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	wg := &sync.WaitGroup{}

	go server.MineBlock(wg)
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

	go gracefulShutdown(register, server)
	if err := register.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func gracefulShutdown(register *grpc.Server, server *server.BitcoinServer) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	<-signalChan
	fmt.Println("Sutting down.")

	register.GracefulStop()

	if err := server.Shutdown(); err != nil {
		fmt.Printf("Shutting down err: %v", err)
	}
	fmt.Println("Sutted down.")

	os.Exit(0)
}
