package model

import (
	"Bitcoin/src/bitcoin/client"
)

const (
	//node
	MaxFailedCount = 10
)

type Node struct {
	Addr   string
	Client client.IBitcoinClient
	Failed int
}

func (node *Node) UpdateState(err error) bool {
	if err != nil {
		if node.Failed < MaxFailedCount {
			node.Failed++
		}
	} else {
		if node.Failed > 0 {
			node.Failed--
		}
	}
	return node.Failed == MaxFailedCount
}
