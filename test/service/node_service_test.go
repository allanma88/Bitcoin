package service

import (
	"Bitcoin/src/bitcoin/client"
	"Bitcoin/src/model"
	"Bitcoin/src/protocol"
	"Bitcoin/src/service"
	"Bitcoin/test"
	"errors"
	"fmt"
	"testing"
)

func Test_SendTx_Check_Nodes(t *testing.T) {
	channels := make(map[string]TxChannel)
	makeclient := func(channel TxChannel) client.IBitcoinClient {
		return &TestBitcoinClient{txChannel: channel}
	}
	nodes := generateNodes(service.MaxBroadcastNodes+5, 5001, channels, makeclient)

	tx := test.NewTransaction([]byte{})
	endpoint := "localhost:5000"
	serv := service.NewNodeService(endpoint, nil)
	serv.AddNodes(nodes...)
	serv.SendTx(tx)

	n, err := checkNodes(channels, endpoint)
	if err != nil {
		t.Fatal(err.Error())
	}
	if n != len(channels) {
		t.Fatalf("expect receive %d request, actual: %d", len(channels), n)
	}
}

func Test_SendTx_Not_Remove_Failed_Not_Enough_Nodes(t *testing.T) {
	makeclient := func(channel TxChannel) client.IBitcoinClient {
		return &TestBitcoinClient{txChannel: channel}
	}
	makeFailedClient := func(channel TxChannel) client.IBitcoinClient {
		return &FailedBitcoinClient{txChannel: channel}
	}

	sendCounts := []int{model.MaxFailedCount - 2, model.MaxFailedCount - 1, model.MaxFailedCount - 5}
	for _, sendCount := range sendCounts {
		channels := make(map[string]TxChannel)
		nodes := generateNodes(service.MaxBroadcastNodes+5, 5001, channels, makeclient)
		failNodes := generateNodes(2, 6001, channels, makeFailedClient)

		endpoint := "localhost:5000"
		serv := service.NewNodeService(endpoint, nil)
		serv.AddNodes(nodes...)
		serv.AddNodes(failNodes...)

		tx := test.NewTransaction([]byte{})
		for i := 0; i < sendCount; i++ {
			serv.SendTx(tx)

			n, err := checkNodes(channels, endpoint)
			if err != nil {
				t.Fatal(err.Error())
			}
			if n != len(channels) {
				t.Fatalf("expect receive %d request, actual: %d", len(channels), n)
			}
		}

		for _, failureNode := range failNodes {
			node := serv.GetNode(failureNode.Addr)
			if node == nil {
				t.Fatalf("failure node %v should not removed since it failed enough", failureNode.Addr)
			}
		}
	}
}

func Test_SendTx_Remove_Failed_Enough_Nodes(t *testing.T) {
	makeclient := func(channel TxChannel) client.IBitcoinClient {
		return &TestBitcoinClient{txChannel: channel}
	}
	makeFailedClient := func(channel TxChannel) client.IBitcoinClient {
		return &FailedBitcoinClient{txChannel: channel}
	}

	sendCounts := []int{model.MaxFailedCount, model.MaxFailedCount + 1, model.MaxFailedCount + 10}
	for _, sendCount := range sendCounts {
		channels := make(map[string]TxChannel)
		nodes := generateNodes(service.MaxBroadcastNodes+5, 5001, channels, makeclient)
		failNodes := generateNodes(2, 6001, channels, makeFailedClient)

		endpoint := "localhost:5000"
		serv := service.NewNodeService(endpoint, nil)
		serv.AddNodes(nodes...)
		serv.AddNodes(failNodes...)

		tx := test.NewTransaction([]byte{})
		for i := 0; i < sendCount; i++ {
			serv.SendTx(tx)

			n, err := checkNodes(channels, endpoint)
			if err != nil {
				t.Fatal(err.Error())
			}

			expect := len(channels)
			if i >= model.MaxFailedCount {
				expect = len(channels) - len(failNodes)
			}

			if n != expect {
				t.Fatalf("expect receive %d request, actual: %d", expect, n)
			}
		}

		for _, failureNode := range failNodes {
			node := serv.GetNode(failureNode.Addr)
			if node != nil {
				t.Fatalf("failure node %v is not removed after it failed enough", failureNode.Addr)
			}
		}
	}
}

func Test_SendTx_Not_Remove_Rarely_Failed_Nodes(t *testing.T) {
	makeclient := func(channel TxChannel) client.IBitcoinClient {
		return &TestBitcoinClient{txChannel: channel}
	}
	makeProbablyFailedClient := func(channel TxChannel) client.IBitcoinClient {
		return &ProbablyFailedBitcoinClient{txChannel: channel, m: 0, n: 2}
	}

	channels := make(map[string]TxChannel)
	nodes := generateNodes(service.MaxBroadcastNodes+5, 5001, channels, makeclient)
	probablyFailNodes := generateNodes(2, 6001, channels, makeProbablyFailedClient)

	endpoint := "localhost:5000"
	serv := service.NewNodeService(endpoint, nil)
	serv.AddNodes(nodes...)
	serv.AddNodes(probablyFailNodes...)

	tx := test.NewTransaction([]byte{})
	for i := 0; i < model.MaxFailedCount*2; i++ {
		serv.SendTx(tx)

		_, err := checkNodes(channels, endpoint)
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	for _, n := range probablyFailNodes {
		node := serv.GetNode(n.Addr)
		if node == nil {
			t.Fatalf("failure node %v should not removed", n.Addr)
		}
	}
}

func Test_RandomPick(t *testing.T) {
	endpoint := "localhost:5000"
	serv := service.NewNodeService(endpoint, nil)
	for i := 0; i < service.MaxBroadcastNodes+5; i++ {
		serv.AddNodes(&model.Node{Addr: fmt.Sprintf("localhost:%d", 5000+i+1)})
	}

	addrs := serv.RandomPickAddrs(service.MaxBroadcastNodes)
	if len(addrs) != service.MaxBroadcastNodes+1 {
		t.Fatalf("expect pick %v nodes, actual: %v", service.MaxBroadcastNodes+1, len(addrs))
	}

	if addrs[0] != endpoint {
		t.Fatalf("the first expect node: %v, actual: %v", endpoint, addrs[0])
	}

	addrMap := make(map[string]string)
	for _, addr := range addrs {
		_, has := addrMap[addr]
		if has {
			t.Errorf("duplicate node: %v", addr)
		}
		addrMap[addr] = addr
	}
}

func Test_Add_Addrs(t *testing.T) {
	serv := service.NewNodeService("localhost:5000", nil)

	addrs := make([]string, 5)
	for i := 0; i < len(addrs); i++ {
		addrs[i] = fmt.Sprintf("localhost:%d", 5001+i)
	}

	err := serv.AddAddrs(addrs)
	if err != nil {
		t.Fatalf("add addrs error: %s", err)
	}

	for i := 0; i < len(addrs); i++ {
		node := serv.GetNode(addrs[i])
		if node == nil {
			t.Fatalf("can not find the node %s", addrs[i])
		}
	}
}

func Test_Add_Invalid_Addrs(t *testing.T) {
	serv := service.NewNodeService("localhost:5000", nil)

	addrs := make([]string, 5)
	for i := 0; i < len(addrs); i++ {
		addrs[i] = fmt.Sprintf("localhost:%d", 5001+i)
	}
	addrs[2] = "127:5001"

	err := serv.AddAddrs(addrs)
	if err == nil {
		t.Fatal("should error for invalid addr")
	}
}

type TxChannel chan *protocol.TransactionReq
type MakeClient func(TxChannel) client.IBitcoinClient

func generateNodes(n, start int, channels map[string]TxChannel, makeclient MakeClient) []*model.Node {
	nodes := make([]*model.Node, n)

	for i := 0; i < n; i++ {
		addr := fmt.Sprintf("localhost:%d", start+i)
		channels[addr] = make(chan *protocol.TransactionReq, 1)
		nodes[i] = &model.Node{
			Addr:   addr,
			Client: makeclient(channels[addr]),
		}
	}

	return nodes
}

func checkNodes(channels map[string]TxChannel, endpoint string) (int, error) {
	n := 0
	for _, channel := range channels {
		select {
		case req := <-channel:
			if len(req.Nodes) != service.MaxBroadcastNodes+1 {
				return n, fmt.Errorf("the nodes size of transaction request are invalid, expect: %v, actual: %v", service.MaxBroadcastNodes+1, len(req.Nodes))
			}

			if req.Nodes[0] != endpoint {
				return n, fmt.Errorf("the first node of transaction request are invalid, expect: %v, actual: %v", endpoint, req.Nodes[0])
			}

			addrMap := make(map[string]string)
			for _, addr := range req.Nodes {
				_, has := channels[addr]
				if !has && addr != endpoint {
					return n, fmt.Errorf("unkown node: %v", addr)
				}

				_, exist := addrMap[addr]
				if exist {
					return n, fmt.Errorf("duplicate node: %v", addr)
				}
				addrMap[addr] = addr
			}
			n++
		default:
		}
	}
	return n, nil
}

type TestBitcoinClient struct {
	txChannel    chan *protocol.TransactionReq
	blockChannel chan *protocol.BlockReq
}

func (client *TestBitcoinClient) SendTx(req *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	client.txChannel <- req
	return &protocol.TransactionReply{Result: true}, nil
}

func (client *TestBitcoinClient) SendBlock(req *protocol.BlockReq) (*protocol.BlockReply, error) {
	// client.channel <- req
	return &protocol.BlockReply{Result: true}, nil
}

func (cli *TestBitcoinClient) GetBlocks(req *protocol.GetBlocksReq) (*protocol.GetBlocksReply, error) {
	return nil, errors.New("not implemented")
}

type FailedBitcoinClient struct {
	txChannel    chan *protocol.TransactionReq
	blockChannel chan *protocol.BlockReq
}

func (client *FailedBitcoinClient) SendTx(req *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	client.txChannel <- req
	return &protocol.TransactionReply{Result: false}, errors.New("send tx failed")
}

func (client *FailedBitcoinClient) SendBlock(req *protocol.BlockReq) (*protocol.BlockReply, error) {
	// client.channel <- req
	return &protocol.BlockReply{Result: false}, errors.New("send tx failed")
}

func (cli *FailedBitcoinClient) GetBlocks(req *protocol.GetBlocksReq) (*protocol.GetBlocksReply, error) {
	return nil, errors.New("not implemented")
}

type ProbablyFailedBitcoinClient struct {
	txChannel    chan *protocol.TransactionReq
	blockChannel chan *protocol.BlockReq
	m            int
	n            int
}

func (client *ProbablyFailedBitcoinClient) SendTx(req *protocol.TransactionReq) (*protocol.TransactionReply, error) {
	client.m++
	if client.m == client.n {
		client.m = 0
		return &protocol.TransactionReply{Result: false}, errors.New("send tx failed")
	} else {
		client.txChannel <- req
		return &protocol.TransactionReply{Result: true}, nil
	}
}

func (client *ProbablyFailedBitcoinClient) SendBlock(req *protocol.BlockReq) (*protocol.BlockReply, error) {
	client.m++
	if client.m == client.n {
		client.m = 0
		return &protocol.BlockReply{Result: false}, errors.New("send tx failed")
	} else {
		// client.channel <- req
		return &protocol.BlockReply{Result: true}, nil
	}
}

func (cli *ProbablyFailedBitcoinClient) GetBlocks(req *protocol.GetBlocksReq) (*protocol.GetBlocksReply, error) {
	return nil, errors.New("not implemented")
}
