package service

import (
	"Bitcoin/src/bitcoin/client"
	"Bitcoin/src/config"
	"Bitcoin/src/model"
	"log"
	"math/rand"
	"net/url"
	"sync"
)

//TODO: maybe we can use more complex policy to remove inactive nodes

type Node struct {
	addr        string
	failedCount int
	client      client.IBitcoinClient
}

type NodeService struct {
	lock  sync.RWMutex
	nodes map[string]*Node
	cfg   *config.Config
}

func NewNodeService(cfg *config.Config) *NodeService {
	service := &NodeService{
		lock:  sync.RWMutex{},
		nodes: make(map[string]*Node),
		cfg:   cfg,
	}
	for _, node := range cfg.Bootstraps {
		service.nodes[node] = &Node{addr: node}
	}
	return service
}

func (service *NodeService) AddNodes(addrs []string) {
	service.lock.Lock()
	for _, addr := range addrs {
		_, err := url.Parse(addr)
		_, has := service.nodes[addr]
		if err == nil && !has {
			service.nodes[addr] = &Node{
				addr:   addr,
				client: client.NewBitcoinClient(addr),
			}
		}
	}
	service.lock.Unlock()
}

func (service *NodeService) SendTx(tx *model.Transaction) {
	nodes := make(map[string]*Node)
	keys := make([]string, 0, len(nodes))

	service.lock.RLock()
	for k, v := range service.nodes {
		keys = append(keys, k)
		nodes[k] = v
	}
	service.lock.RUnlock()

	addrs := service.randomPick(keys, model.MaxBroadcastNodes)
	deleted := make([]string, 0, len(nodes))

	for addr, node := range nodes {
		req := model.TransactionTo(tx, addrs)

		reply, err := node.client.SendTx(req)
		if err != nil {
			log.Printf("sent transaction failed: %v", err)
			node.failedCount++
			if node.failedCount > model.MaxFailedCount {
				deleted = append(deleted, addr)
			}
		}
		log.Printf("sent transaction result: %v", reply.Result)
	}

	service.lock.Lock()
	for _, node := range deleted {
		delete(service.nodes, node)
	}
	service.lock.Unlock()
}

func (service *NodeService) randomPick(addrs []string, n int) []string {
	if n > len(addrs) {
		n = len(addrs)
	}
	indics := rand.Perm(len(addrs))
	selects := make([]string, 1, n+1)
	selects[0] = service.cfg.Endpoint
	for i := 0; i < n; i++ {
		selects = append(selects, addrs[indics[i]])
	}
	return selects
}
