package model

import "Bitcoin/src/bitcoin/client"

type Node struct {
	Addr   string
	Client client.IBitcoinClient
	Failed int
}
