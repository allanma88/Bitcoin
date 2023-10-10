package service

import (
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"context"
	"log"
	"math/rand"
	"net/url"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

//TODO: maybe we can use more complex policy to remove inactive nodes

type NodeService struct {
	lock  sync.RWMutex
	nodes map[string]*model.Node
}

func NewNodeService() *NodeService {
	service := &NodeService{
		lock:  sync.RWMutex{},
		nodes: make(map[string]*model.Node),
	}
	return service
}

func (service *NodeService) AddNodes(addrs []string) {
	service.lock.Lock()
	for _, addr := range addrs {
		_, err := url.Parse(addr)
		_, has := service.nodes[addr]
		if err == nil && !has {
			service.nodes[addr] = &model.Node{Addr: addr}
		}
	}
	service.lock.Unlock()
}

func (service *NodeService) SendTx(tx *model.Transaction) {
	nodes := make(map[string]*model.Node)
	keys := make([]string, 0, len(nodes))

	service.lock.RLock()
	for k, v := range service.nodes {
		keys = append(keys, k)
		nodes[k] = v
	}
	service.lock.RUnlock()

	addrs := randomPick(keys, model.MaxBroadcastNodes)
	deleted := make([]string, 0, len(nodes))

	for addr, node := range nodes {
		err := service.sendTx(node, tx, addrs)
		if err != nil {
			log.Printf("sent transaction failed: %v", err)
			node.FailedCount++
			if node.FailedCount > model.MaxFailedCount {
				deleted = append(deleted, addr)
			}
		}
	}

	service.lock.Lock()
	for _, node := range deleted {
		delete(service.nodes, node)
	}
	service.lock.Unlock()
}

func (service *NodeService) sendTx(node *model.Node, tx *model.Transaction, addrs []string) error {
	//TODO: not finished
	if node.Connection == nil || node.Connection.GetState() == connectivity.Shutdown {
		// Set up a connection to the server.
		conn, err := grpc.Dial(node.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		log.Printf("connected to %s", node.Addr)
		node.Connection = conn
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := protocol.NewTransactionClient(node.Connection)
	req := model.TransactionTo(tx, addrs)
	reply, err := client.ExecuteTx(ctx, req)
	if err != nil {
		return err
	}
	log.Printf("sent transaction result: %v", reply.Result)
	return nil
}

func randomPick(addrs []string, n int) []string {
	if n > len(addrs) {
		n = len(addrs)
	}
	indics := rand.Perm(len(addrs))
	selects := make([]string, 0, n)
	for i := 0; i < n; i++ {
		selects = append(selects, addrs[indics[i]])
	}
	return selects
}
