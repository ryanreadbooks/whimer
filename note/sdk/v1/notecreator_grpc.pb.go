// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: v1/notecreator.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	NoteCreatorService_IsUserOwnNote_FullMethodName = "/note.sdk.v1.NoteCreatorService/IsUserOwnNote"
	NoteCreatorService_IsNoteExist_FullMethodName   = "/note.sdk.v1.NoteCreatorService/IsNoteExist"
	NoteCreatorService_CreateNote_FullMethodName    = "/note.sdk.v1.NoteCreatorService/CreateNote"
	NoteCreatorService_UpdateNote_FullMethodName    = "/note.sdk.v1.NoteCreatorService/UpdateNote"
	NoteCreatorService_DeleteNote_FullMethodName    = "/note.sdk.v1.NoteCreatorService/DeleteNote"
	NoteCreatorService_GetNote_FullMethodName       = "/note.sdk.v1.NoteCreatorService/GetNote"
	NoteCreatorService_ListNote_FullMethodName      = "/note.sdk.v1.NoteCreatorService/ListNote"
	NoteCreatorService_GetUploadAuth_FullMethodName = "/note.sdk.v1.NoteCreatorService/GetUploadAuth"
)

// NoteCreatorServiceClient is the client API for NoteCreatorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// 和笔记管理相关的服务
// 比如发布笔记，修改笔记，删除笔记等管理笔记的功能
type NoteCreatorServiceClient interface {
	// 检查用户是否拥有指定的笔记
	IsUserOwnNote(ctx context.Context, in *IsUserOwnNoteRequest, opts ...grpc.CallOption) (*IsUserOwnNoteResponse, error)
	// 判断笔记是否存在
	IsNoteExist(ctx context.Context, in *IsNoteExistRequest, opts ...grpc.CallOption) (*IsNoteExistResponse, error)
	// 创建笔记
	CreateNote(ctx context.Context, in *CreateNoteRequest, opts ...grpc.CallOption) (*CreateNoteResponse, error)
	// 更新笔记
	UpdateNote(ctx context.Context, in *UpdateNoteRequest, opts ...grpc.CallOption) (*UpdateNoteResponse, error)
	// 删除笔记
	DeleteNote(ctx context.Context, in *DeleteNoteRequest, opts ...grpc.CallOption) (*DeleteNoteResponse, error)
	// 获取笔记的信息
	GetNote(ctx context.Context, in *GetNoteRequest, opts ...grpc.CallOption) (*GetNoteResponse, error)
	// 列出笔记
	ListNote(ctx context.Context, in *ListNoteRequest, opts ...grpc.CallOption) (*ListNoteResponse, error)
	// 获取上传凭证
	GetUploadAuth(ctx context.Context, in *GetUploadAuthRequest, opts ...grpc.CallOption) (*GetUploadAuthResponse, error)
}

type noteCreatorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNoteCreatorServiceClient(cc grpc.ClientConnInterface) NoteCreatorServiceClient {
	return &noteCreatorServiceClient{cc}
}

func (c *noteCreatorServiceClient) IsUserOwnNote(ctx context.Context, in *IsUserOwnNoteRequest, opts ...grpc.CallOption) (*IsUserOwnNoteResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IsUserOwnNoteResponse)
	err := c.cc.Invoke(ctx, NoteCreatorService_IsUserOwnNote_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *noteCreatorServiceClient) IsNoteExist(ctx context.Context, in *IsNoteExistRequest, opts ...grpc.CallOption) (*IsNoteExistResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(IsNoteExistResponse)
	err := c.cc.Invoke(ctx, NoteCreatorService_IsNoteExist_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *noteCreatorServiceClient) CreateNote(ctx context.Context, in *CreateNoteRequest, opts ...grpc.CallOption) (*CreateNoteResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateNoteResponse)
	err := c.cc.Invoke(ctx, NoteCreatorService_CreateNote_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *noteCreatorServiceClient) UpdateNote(ctx context.Context, in *UpdateNoteRequest, opts ...grpc.CallOption) (*UpdateNoteResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateNoteResponse)
	err := c.cc.Invoke(ctx, NoteCreatorService_UpdateNote_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *noteCreatorServiceClient) DeleteNote(ctx context.Context, in *DeleteNoteRequest, opts ...grpc.CallOption) (*DeleteNoteResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteNoteResponse)
	err := c.cc.Invoke(ctx, NoteCreatorService_DeleteNote_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *noteCreatorServiceClient) GetNote(ctx context.Context, in *GetNoteRequest, opts ...grpc.CallOption) (*GetNoteResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetNoteResponse)
	err := c.cc.Invoke(ctx, NoteCreatorService_GetNote_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *noteCreatorServiceClient) ListNote(ctx context.Context, in *ListNoteRequest, opts ...grpc.CallOption) (*ListNoteResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListNoteResponse)
	err := c.cc.Invoke(ctx, NoteCreatorService_ListNote_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *noteCreatorServiceClient) GetUploadAuth(ctx context.Context, in *GetUploadAuthRequest, opts ...grpc.CallOption) (*GetUploadAuthResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetUploadAuthResponse)
	err := c.cc.Invoke(ctx, NoteCreatorService_GetUploadAuth_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NoteCreatorServiceServer is the server API for NoteCreatorService service.
// All implementations must embed UnimplementedNoteCreatorServiceServer
// for forward compatibility.
//
// 和笔记管理相关的服务
// 比如发布笔记，修改笔记，删除笔记等管理笔记的功能
type NoteCreatorServiceServer interface {
	// 检查用户是否拥有指定的笔记
	IsUserOwnNote(context.Context, *IsUserOwnNoteRequest) (*IsUserOwnNoteResponse, error)
	// 判断笔记是否存在
	IsNoteExist(context.Context, *IsNoteExistRequest) (*IsNoteExistResponse, error)
	// 创建笔记
	CreateNote(context.Context, *CreateNoteRequest) (*CreateNoteResponse, error)
	// 更新笔记
	UpdateNote(context.Context, *UpdateNoteRequest) (*UpdateNoteResponse, error)
	// 删除笔记
	DeleteNote(context.Context, *DeleteNoteRequest) (*DeleteNoteResponse, error)
	// 获取笔记的信息
	GetNote(context.Context, *GetNoteRequest) (*GetNoteResponse, error)
	// 列出笔记
	ListNote(context.Context, *ListNoteRequest) (*ListNoteResponse, error)
	// 获取上传凭证
	GetUploadAuth(context.Context, *GetUploadAuthRequest) (*GetUploadAuthResponse, error)
	mustEmbedUnimplementedNoteCreatorServiceServer()
}

// UnimplementedNoteCreatorServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedNoteCreatorServiceServer struct{}

func (UnimplementedNoteCreatorServiceServer) IsUserOwnNote(context.Context, *IsUserOwnNoteRequest) (*IsUserOwnNoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsUserOwnNote not implemented")
}
func (UnimplementedNoteCreatorServiceServer) IsNoteExist(context.Context, *IsNoteExistRequest) (*IsNoteExistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsNoteExist not implemented")
}
func (UnimplementedNoteCreatorServiceServer) CreateNote(context.Context, *CreateNoteRequest) (*CreateNoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateNote not implemented")
}
func (UnimplementedNoteCreatorServiceServer) UpdateNote(context.Context, *UpdateNoteRequest) (*UpdateNoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateNote not implemented")
}
func (UnimplementedNoteCreatorServiceServer) DeleteNote(context.Context, *DeleteNoteRequest) (*DeleteNoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNote not implemented")
}
func (UnimplementedNoteCreatorServiceServer) GetNote(context.Context, *GetNoteRequest) (*GetNoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNote not implemented")
}
func (UnimplementedNoteCreatorServiceServer) ListNote(context.Context, *ListNoteRequest) (*ListNoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListNote not implemented")
}
func (UnimplementedNoteCreatorServiceServer) GetUploadAuth(context.Context, *GetUploadAuthRequest) (*GetUploadAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUploadAuth not implemented")
}
func (UnimplementedNoteCreatorServiceServer) mustEmbedUnimplementedNoteCreatorServiceServer() {}
func (UnimplementedNoteCreatorServiceServer) testEmbeddedByValue()                            {}

// UnsafeNoteCreatorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NoteCreatorServiceServer will
// result in compilation errors.
type UnsafeNoteCreatorServiceServer interface {
	mustEmbedUnimplementedNoteCreatorServiceServer()
}

func RegisterNoteCreatorServiceServer(s grpc.ServiceRegistrar, srv NoteCreatorServiceServer) {
	// If the following call pancis, it indicates UnimplementedNoteCreatorServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&NoteCreatorService_ServiceDesc, srv)
}

func _NoteCreatorService_IsUserOwnNote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsUserOwnNoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteCreatorServiceServer).IsUserOwnNote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NoteCreatorService_IsUserOwnNote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteCreatorServiceServer).IsUserOwnNote(ctx, req.(*IsUserOwnNoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NoteCreatorService_IsNoteExist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsNoteExistRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteCreatorServiceServer).IsNoteExist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NoteCreatorService_IsNoteExist_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteCreatorServiceServer).IsNoteExist(ctx, req.(*IsNoteExistRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NoteCreatorService_CreateNote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateNoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteCreatorServiceServer).CreateNote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NoteCreatorService_CreateNote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteCreatorServiceServer).CreateNote(ctx, req.(*CreateNoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NoteCreatorService_UpdateNote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateNoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteCreatorServiceServer).UpdateNote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NoteCreatorService_UpdateNote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteCreatorServiceServer).UpdateNote(ctx, req.(*UpdateNoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NoteCreatorService_DeleteNote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteNoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteCreatorServiceServer).DeleteNote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NoteCreatorService_DeleteNote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteCreatorServiceServer).DeleteNote(ctx, req.(*DeleteNoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NoteCreatorService_GetNote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetNoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteCreatorServiceServer).GetNote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NoteCreatorService_GetNote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteCreatorServiceServer).GetNote(ctx, req.(*GetNoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NoteCreatorService_ListNote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListNoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteCreatorServiceServer).ListNote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NoteCreatorService_ListNote_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteCreatorServiceServer).ListNote(ctx, req.(*ListNoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NoteCreatorService_GetUploadAuth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUploadAuthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NoteCreatorServiceServer).GetUploadAuth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NoteCreatorService_GetUploadAuth_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NoteCreatorServiceServer).GetUploadAuth(ctx, req.(*GetUploadAuthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NoteCreatorService_ServiceDesc is the grpc.ServiceDesc for NoteCreatorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NoteCreatorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "note.sdk.v1.NoteCreatorService",
	HandlerType: (*NoteCreatorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "IsUserOwnNote",
			Handler:    _NoteCreatorService_IsUserOwnNote_Handler,
		},
		{
			MethodName: "IsNoteExist",
			Handler:    _NoteCreatorService_IsNoteExist_Handler,
		},
		{
			MethodName: "CreateNote",
			Handler:    _NoteCreatorService_CreateNote_Handler,
		},
		{
			MethodName: "UpdateNote",
			Handler:    _NoteCreatorService_UpdateNote_Handler,
		},
		{
			MethodName: "DeleteNote",
			Handler:    _NoteCreatorService_DeleteNote_Handler,
		},
		{
			MethodName: "GetNote",
			Handler:    _NoteCreatorService_GetNote_Handler,
		},
		{
			MethodName: "ListNote",
			Handler:    _NoteCreatorService_ListNote_Handler,
		},
		{
			MethodName: "GetUploadAuth",
			Handler:    _NoteCreatorService_GetUploadAuth_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/notecreator.proto",
}
