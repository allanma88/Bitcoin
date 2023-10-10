package model

import (
	"google.golang.org/grpc"
)

type Node struct {
	Addr        string
	FailedCount int
	Connection  *grpc.ClientConn
}
