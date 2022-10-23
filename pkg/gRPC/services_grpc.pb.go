// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: services.proto

package grpc

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

// DutchServiceClient is the client API for DutchService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DutchServiceClient interface {
	SignUp(ctx context.Context, in *SignUpRequest, opts ...grpc.CallOption) (*SignUpResponse, error)
	SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*SignInResponse, error)
	SignOut(ctx context.Context, in *SignOutRequest, opts ...grpc.CallOption) (*SignOutResponse, error)
}

type dutchServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDutchServiceClient(cc grpc.ClientConnInterface) DutchServiceClient {
	return &dutchServiceClient{cc}
}

func (c *dutchServiceClient) SignUp(ctx context.Context, in *SignUpRequest, opts ...grpc.CallOption) (*SignUpResponse, error) {
	out := new(SignUpResponse)
	err := c.cc.Invoke(ctx, "/DutchService/SignUp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dutchServiceClient) SignIn(ctx context.Context, in *SignInRequest, opts ...grpc.CallOption) (*SignInResponse, error) {
	out := new(SignInResponse)
	err := c.cc.Invoke(ctx, "/DutchService/SignIn", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dutchServiceClient) SignOut(ctx context.Context, in *SignOutRequest, opts ...grpc.CallOption) (*SignOutResponse, error) {
	out := new(SignOutResponse)
	err := c.cc.Invoke(ctx, "/DutchService/SignOut", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DutchServiceServer is the server API for DutchService service.
// All implementations must embed UnimplementedDutchServiceServer
// for forward compatibility
type DutchServiceServer interface {
	SignUp(context.Context, *SignUpRequest) (*SignUpResponse, error)
	SignIn(context.Context, *SignInRequest) (*SignInResponse, error)
	SignOut(context.Context, *SignOutRequest) (*SignOutResponse, error)
	mustEmbedUnimplementedDutchServiceServer()
}

// UnimplementedDutchServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDutchServiceServer struct {
}

func (UnimplementedDutchServiceServer) SignUp(context.Context, *SignUpRequest) (*SignUpResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignUp not implemented")
}
func (UnimplementedDutchServiceServer) SignIn(context.Context, *SignInRequest) (*SignInResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignIn not implemented")
}
func (UnimplementedDutchServiceServer) SignOut(context.Context, *SignOutRequest) (*SignOutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignOut not implemented")
}
func (UnimplementedDutchServiceServer) mustEmbedUnimplementedDutchServiceServer() {}

// UnsafeDutchServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DutchServiceServer will
// result in compilation errors.
type UnsafeDutchServiceServer interface {
	mustEmbedUnimplementedDutchServiceServer()
}

func RegisterDutchServiceServer(s grpc.ServiceRegistrar, srv DutchServiceServer) {
	s.RegisterService(&DutchService_ServiceDesc, srv)
}

func _DutchService_SignUp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignUpRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DutchServiceServer).SignUp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DutchService/SignUp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DutchServiceServer).SignUp(ctx, req.(*SignUpRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DutchService_SignIn_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignInRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DutchServiceServer).SignIn(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DutchService/SignIn",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DutchServiceServer).SignIn(ctx, req.(*SignInRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DutchService_SignOut_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignOutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DutchServiceServer).SignOut(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DutchService/SignOut",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DutchServiceServer).SignOut(ctx, req.(*SignOutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DutchService_ServiceDesc is the grpc.ServiceDesc for DutchService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DutchService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DutchService",
	HandlerType: (*DutchServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SignUp",
			Handler:    _DutchService_SignUp_Handler,
		},
		{
			MethodName: "SignIn",
			Handler:    _DutchService_SignIn_Handler,
		},
		{
			MethodName: "SignOut",
			Handler:    _DutchService_SignOut_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "services.proto",
}
