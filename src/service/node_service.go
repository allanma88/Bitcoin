package service

import (
	"Bitcoin/src/bitcoin/client"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sync"
)

const (
	MaxBroadcastNodes = 10
)

//TODO: maybe we can use more complex policy to remove inactive nodes

type NodeService struct {
	lock     sync.RWMutex
	nodes    map[string]*model.Node
	endpoint string
}

type sendFunc[Q, R any] func(cli client.IBitcoinClient, req Q) (R, error)

func NewNodeService(endpoint string, bootstraps []string) *NodeService {
	service := &NodeService{
		lock:     sync.RWMutex{},
		nodes:    make(map[string]*model.Node),
		endpoint: endpoint,
	}
	if bootstraps != nil {
		for _, addr := range bootstraps {
			service.nodes[addr] = &model.Node{Addr: addr}
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
		service.nodes[node.Addr] = node
	}
	return nil
}

// TODO: remove?
func (service *NodeService) GetNode(addr string) *model.Node {
	return service.nodes[addr]
}

func (service *NodeService) SendTx(tx *model.Transaction) {
	send := func(cli client.IBitcoinClient, req *protocol.TransactionReq) (*protocol.TransactionReply, error) {
		return cli.SendTx(req)
	}

	req := model.TransactionTo(tx)
	req.Nodes = service.RandomPickAddrs(MaxBroadcastNodes)

	broadcastReq[*protocol.TransactionReq](service, req, send)
}

func (service *NodeService) SendBlock(block *model.Block) {
	send := func(cli client.IBitcoinClient, req *protocol.BlockReq) (*protocol.BlockReply, error) {
		return cli.SendBlock(req)
	}

	blockReq, err := model.BlockTo(block)
	if err != nil {
		log.Printf("convert to block request error: %v", err)
		return
	}

	broadcastReq[*protocol.BlockReq](service, blockReq, send)
}

// TODO: test cases
func (service *NodeService) GetBlocks(blockHashes [][]byte, addr string) ([]*protocol.BlockReq, uint64, error) {
	req := &protocol.GetBlocksReq{
		Blockhashes: blockHashes,
	}
	node := service.nodes[addr]
	reply, err := node.Client.GetBlocks(req)
	removed := node.UpdateState(err)

	if removed {
		service.lock.Lock()
		delete(service.nodes, addr)
		service.lock.Unlock()
	}
	if err != nil {
		return nil, 0, err
	}
	return reply.Blocks, reply.End, nil
}

func broadcastReq[Q, R any](service *NodeService, req Q, send sendFunc[Q, R]) {
	deleted := make([]string, 0, len(service.nodes))
	wg := &sync.WaitGroup{}

	nodes := service.RandomPickNodes(MaxBroadcastNodes)
	for _, node := range nodes {
		wg.Add(1)
		go func(n *model.Node) {
			_, err := send(n.Client, req)
			removed := n.UpdateState(err)
			if removed {
				deleted = append(deleted, n.Addr)
			}
			wg.Done()
		}(node)
	}
	wg.Wait()

	service.lock.Lock()
	for _, node := range deleted {
		delete(service.nodes, node)
	}
	service.lock.Unlock()
}

func (service *NodeService) RandomPickAddrs(n int) []string {
	service.lock.RLock()
	addrs := make([]string, 0, len(service.nodes))
	for k := range service.nodes {
		addrs = append(addrs, k)
	}
	if n > len(addrs) {
		n = len(addrs)
	}
	service.lock.RUnlock()

	indics := rand.Perm(n)
	selects := make([]string, 1, n+1)
	selects[0] = service.endpoint
	for i := 0; i < n; i++ {
		selects = append(selects, addrs[indics[i]])
	}
	return selects
}

func (service *NodeService) RandomPickNodes(n int) []*model.Node {
	service.lock.RLock()
	nodes := make([]*model.Node, 0, len(service.nodes))
	for _, node := range service.nodes {
		nodes = append(nodes, node)
	}
	if n > len(nodes) {
		n = len(nodes)
	}
	service.lock.RUnlock()

	indics := rand.Perm(n)
	selects := make([]*model.Node, 0, n)
	for i := 0; i < n; i++ {
		selects = append(selects, nodes[indics[i]])
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
