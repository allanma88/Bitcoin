package service

import (
	"Bitcoin/src/bitcoin/client"
	"Bitcoin/src/config"
	"Bitcoin/src/model"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sync"
)

//TODO: maybe we can use more complex policy to remove inactive nodes

type Node struct {
	Addr   string
	Client client.IBitcoinClient
	failed int
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
	if cfg.Bootstraps != nil {
		for _, node := range cfg.Bootstraps {
			service.nodes[node] = &Node{Addr: node}
		}
	}
	return service
}

func (service *NodeService) AddAddrs(addrs []string) error {
	nodes, err := toNodes(addrs)
	if err != nil {
		return err
	}

	return service.AddNodes(nodes...)
}

func (service *NodeService) AddNodes(nodes ...*Node) error {
	service.lock.Lock()
	defer service.lock.Unlock()

	for _, node := range nodes {
		if node == nil {
			log.Fatalf("node is nil")
		}
		_, has := service.nodes[node.Addr]
		if has {
			return fmt.Errorf("the node %s already exists", node.Addr)
		}
		service.nodes[node.Addr] = node
	}
	return nil
}

func (service *NodeService) GetNode(addr string) *Node {
	return service.nodes[addr]
}

func (service *NodeService) SendTx(tx *model.Transaction) {
	nodes := make(map[string]*Node)
	addrs := make([]string, 0, len(nodes))

	service.lock.RLock()
	for k, v := range service.nodes {
		addrs = append(addrs, k)
		nodes[k] = v
	}
	service.lock.RUnlock()

	selectedAddrs := RandomPick(service.cfg.Endpoint, addrs, model.MaxBroadcastNodes)
	deleted := make([]string, 0, len(nodes))

	for addr, node := range nodes {
		req := model.TransactionTo(tx, selectedAddrs)

		_, err := node.Client.SendTx(req)
		if err != nil {
			log.Printf("sent transaction failed: %v", err)
			node.failed++
			if node.failed >= model.MaxFailedCount {
				deleted = append(deleted, addr)
			}
		} else {
			if node.failed > 0 {
				node.failed--
			}
		}
		// log.Printf("sent transaction result: %v", reply.Result)
	}

	service.lock.Lock()
	for _, node := range deleted {
		delete(service.nodes, node)
	}
	service.lock.Unlock()
}

func RandomPick(endpoint string, addrs []string, n int) []string {
	if n > len(addrs) {
		n = len(addrs)
	}
	indics := rand.Perm(len(addrs))
	selects := make([]string, 1, n+1)
	selects[0] = endpoint
	for i := 0; i < n; i++ {
		selects = append(selects, addrs[indics[i]])
	}
	return selects
}

func toNodes(addrs []string) ([]*Node, error) {
	nodes := make([]*Node, len(addrs))

	for i, addr := range addrs {
		_, err := url.Parse(addr)
		if err != nil {
			//TODO: more specific type of error
			return nil, fmt.Errorf("the format of node %s error: %s", addr, err)
		}

		node := &Node{
			Addr:   addr,
			Client: client.NewBitcoinClient(addr),
		}
		nodes[i] = node
	}
	return nodes, nil
}
