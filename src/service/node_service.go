package service

import (
	"Bitcoin/src/bitcoin/client"
	"Bitcoin/src/config"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sync"
)

//TODO: maybe we can use more complex policy to remove inactive nodes

type NodeService struct {
	lock  sync.RWMutex
	nodes map[string]*model.Node
	cfg   *config.Config
}

type sendFunc[T any] func(cli client.IBitcoinClient, req T) error

func NewNodeService(cfg *config.Config) *NodeService {
	service := &NodeService{
		lock:  sync.RWMutex{},
		nodes: make(map[string]*model.Node),
		cfg:   cfg,
	}
	if cfg.Bootstraps != nil {
		for _, node := range cfg.Bootstraps {
			service.nodes[node] = &model.Node{Addr: node}
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

func (service *NodeService) AddNodes(nodes ...*model.Node) error {
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

func (service *NodeService) GetNode(addr string) *model.Node {
	return service.nodes[addr]
}

func (service *NodeService) SendTx(tx *model.Transaction) {
	addrs := service.RandomPick(model.MaxBroadcastNodes)
	req := model.TransactionTo(tx)
	req.Nodes = addrs

	send := func(cli client.IBitcoinClient, req *protocol.TransactionReq) error {
		_, err := cli.SendTx(req)
		return err
	}

	sendReq[*protocol.TransactionReq](service, req, send)
}

func (service *NodeService) SendBlock(block *model.Block) {
	addrs := service.RandomPick(model.MaxBroadcastNodes)
	req, err := model.BlockTo(block)
	if err != nil {
		log.Printf("convert to block request error: %v", err)
		return
	}
	req.Nodes = addrs

	send := func(cli client.IBitcoinClient, req *protocol.BlockReq) error {
		_, err := cli.SendBlock(req)
		return err
	}

	sendReq[*protocol.BlockReq](service, req, send)
}

func sendReq[T any](service *NodeService, req T, send sendFunc[T]) {
	deleted := make([]string, 0, len(service.nodes))
	wg := &sync.WaitGroup{}

	for _, node := range service.nodes {
		wg.Add(1)
		go func(n *model.Node) {
			err := send(n.Client, req)
			if err != nil {
				log.Printf("sent transaction failed: %v", err)
				n.Failed++
				if n.Failed >= model.MaxFailedCount {
					deleted = append(deleted, n.Addr)
				}
			} else {
				if n.Failed > 0 {
					n.Failed--
				}
			}
			wg.Done()
			// log.Printf("sent transaction result: %v", reply.Result)
		}(node)
	}
	wg.Wait()

	service.lock.Lock()
	for _, node := range deleted {
		delete(service.nodes, node)
	}
	service.lock.Unlock()
}

func (service *NodeService) RandomPick(n int) []string {
	addrs := make([]string, 0, len(service.nodes))

	if n > len(service.nodes) {
		n = len(addrs)
	}

	service.lock.RLock()
	for k := range service.nodes {
		addrs = append(addrs, k)
	}
	service.lock.RUnlock()

	indics := rand.Perm(n)
	selects := make([]string, 1, n+1)
	selects[0] = service.cfg.Endpoint
	for i := 0; i < n; i++ {
		selects = append(selects, addrs[indics[i]])
	}
	return selects
}

func toNodes(addrs []string) ([]*model.Node, error) {
	nodes := make([]*model.Node, len(addrs))

	for i, addr := range addrs {
		_, err := url.Parse(addr)
		if err != nil {
			//TODO: more specific type of error
			return nil, fmt.Errorf("the format of node %s error: %s", addr, err)
		}

		node := &model.Node{
			Addr:   addr,
			Client: client.NewBitcoinClient(addr),
		}
		nodes[i] = node
	}
	return nodes, nil
}
