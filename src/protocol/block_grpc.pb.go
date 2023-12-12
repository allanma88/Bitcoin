// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: block.proto

package protocol

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// BlockClient is the client API for Block service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlockClient interface {
	// find a new block
	NewBlock(ctx context.Context, in *BlockReq, opts ...grpc.CallOption) (*BlockReply, error)
	GetBlocks(ctx context.Context, in *GetBlocksReq, opts ...grpc.CallOption) (*GetBlocksReply, error)
}

type blockClient struct {
	cc grpc.ClientConnInterface
}

func NewBlockClient(cc grpc.ClientConnInterface) BlockClient {
	return &blockClient{cc}
}

func (c *blockClient) NewBlock(ctx context.Context, in *BlockReq, opts ...grpc.CallOption) (*BlockReply, error) {
	out := new(BlockReply)
	err := c.cc.Invoke(ctx, "/protocol.Block/NewBlock", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockClient) GetBlocks(ctx context.Context, in *GetBlocksReq, opts ...grpc.CallOption) (*GetBlocksReply, error) {
	out := new(GetBlocksReply)
	err := c.cc.Invoke(ctx, "/protocol.Block/GetBlocks", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlockServer is the server API for Block service.
// All implementations must embed UnimplementedBlockServer
// for forward compatibility
type BlockServer interface {
	// find a new block
	NewBlock(context.Context, *BlockReq) (*BlockReply, error)
	GetBlocks(context.Context, *GetBlocksReq) (*GetBlocksReply, error)
	mustEmbedUnimplementedBlockServer()
}

// UnimplementedBlockServer must be embedded to have forward compatible implementations.
type UnimplementedBlockServer struct {
}

func (UnimplementedBlockServer) NewBlock(context.Context, *BlockReq) (*BlockReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewBlock not implemented")
}
func (UnimplementedBlockServer) GetBlocks(context.Context, *GetBlocksReq) (*GetBlocksReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlocks not implemented")
}
func (UnimplementedBlockServer) mustEmbedUnimplementedBlockServer() {}

// UnsafeBlockServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BlockServer will
// result in compilation errors.
type UnsafeBlockServer interface {
	mustEmbedUnimplementedBlockServer()
}

func RegisterBlockServer(s grpc.ServiceRegistrar, srv BlockServer) {
	s.RegisterService(&Block_ServiceDesc, srv)
}

func _Block_NewBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockServer).NewBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.Block/NewBlock",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockServer).NewBlock(ctx, req.(*BlockReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Block_GetBlocks_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlocksReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockServer).GetBlocks(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protocol.Block/GetBlocks",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockServer).GetBlocks(ctx, req.(*GetBlocksReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Block_ServiceDesc is the grpc.ServiceDesc for Block service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Block_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protocol.Block",
	HandlerType: (*BlockServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewBlock",
			Handler:    _Block_NewBlock_Handler,
		},
		{
			MethodName: "GetBlocks",
			Handler:    _Block_GetBlocks_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "block.proto",
}
