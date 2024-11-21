// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: v1/reply.proto

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
	ReplyService_AddReply_FullMethodName               = "/comment.sdk.v1.ReplyService/AddReply"
	ReplyService_DelReply_FullMethodName               = "/comment.sdk.v1.ReplyService/DelReply"
	ReplyService_LikeAction_FullMethodName             = "/comment.sdk.v1.ReplyService/LikeAction"
	ReplyService_DislikeAction_FullMethodName          = "/comment.sdk.v1.ReplyService/DislikeAction"
	ReplyService_ReportReply_FullMethodName            = "/comment.sdk.v1.ReplyService/ReportReply"
	ReplyService_PinReply_FullMethodName               = "/comment.sdk.v1.ReplyService/PinReply"
	ReplyService_PageGetReply_FullMethodName           = "/comment.sdk.v1.ReplyService/PageGetReply"
	ReplyService_PageGetSubReply_FullMethodName        = "/comment.sdk.v1.ReplyService/PageGetSubReply"
	ReplyService_PageGetDetailedReply_FullMethodName   = "/comment.sdk.v1.ReplyService/PageGetDetailedReply"
	ReplyService_GetPinnedReply_FullMethodName         = "/comment.sdk.v1.ReplyService/GetPinnedReply"
	ReplyService_CountReply_FullMethodName             = "/comment.sdk.v1.ReplyService/CountReply"
	ReplyService_BatchCountReply_FullMethodName        = "/comment.sdk.v1.ReplyService/BatchCountReply"
	ReplyService_GetReplyLikeCount_FullMethodName      = "/comment.sdk.v1.ReplyService/GetReplyLikeCount"
	ReplyService_GetReplyDislikeCount_FullMethodName   = "/comment.sdk.v1.ReplyService/GetReplyDislikeCount"
	ReplyService_CheckUserOnOjbect_FullMethodName      = "/comment.sdk.v1.ReplyService/CheckUserOnOjbect"
	ReplyService_BatchCheckUserOnOjbect_FullMethodName = "/comment.sdk.v1.ReplyService/BatchCheckUserOnOjbect"
)

// ReplyServiceClient is the client API for ReplyService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReplyServiceClient interface {
	// 发表评论
	AddReply(ctx context.Context, in *AddReplyReq, opts ...grpc.CallOption) (*AddReplyRes, error)
	// 删除评论
	DelReply(ctx context.Context, in *DelReplyReq, opts ...grpc.CallOption) (*DelReplyRes, error)
	// 赞
	LikeAction(ctx context.Context, in *LikeActionReq, opts ...grpc.CallOption) (*LikeActionRes, error)
	// 踩
	DislikeAction(ctx context.Context, in *DislikeActionReq, opts ...grpc.CallOption) (*DislikeActionRes, error)
	// 举报
	ReportReply(ctx context.Context, in *ReportReplyReq, opts ...grpc.CallOption) (*ReportReplyRes, error)
	// 置顶评论
	PinReply(ctx context.Context, in *PinReplyReq, opts ...grpc.CallOption) (*PinReplyRes, error)
	// 获取主评论信息
	PageGetReply(ctx context.Context, in *PageGetReplyReq, opts ...grpc.CallOption) (*PageGetReplyRes, error)
	// 获取子评论信息
	PageGetSubReply(ctx context.Context, in *PageGetSubReplyReq, opts ...grpc.CallOption) (*PageGetSubReplyRes, error)
	// 获取主评论详细信息
	PageGetDetailedReply(ctx context.Context, in *PageGetReplyReq, opts ...grpc.CallOption) (*PageGetDetailedReplyRes, error)
	// 获取置顶评论
	GetPinnedReply(ctx context.Context, in *GetPinnedReplyReq, opts ...grpc.CallOption) (*GetPinnedReplyRes, error)
	// 获取某个被评对象的评论数
	CountReply(ctx context.Context, in *CountReplyReq, opts ...grpc.CallOption) (*CountReplyRes, error)
	// 获取多个被评论对象的评论数
	BatchCountReply(ctx context.Context, in *BatchCountReplyRequest, opts ...grpc.CallOption) (*BatchCountReplyResponse, error)
	// 获取某条评论的点赞数
	GetReplyLikeCount(ctx context.Context, in *GetReplyLikeCountReq, opts ...grpc.CallOption) (*GetReplyLikeCountRes, error)
	// 获取某条评论的点踩数
	GetReplyDislikeCount(ctx context.Context, in *GetReplyDislikeCountReq, opts ...grpc.CallOption) (*GetReplyDislikeCountRes, error)
	// 获取某个用户是否评论了某个对象
	CheckUserOnOjbect(ctx context.Context, in *CheckUserOnOjbectRequest, opts ...grpc.CallOption) (*CheckUserOnOjbectResponse, error)
	BatchCheckUserOnOjbect(ctx context.Context, in *BatchCheckUserOnOjbectRequest, opts ...grpc.CallOption) (*BatchCheckUserOnOjbectResponse, error)
}

type replyServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewReplyServiceClient(cc grpc.ClientConnInterface) ReplyServiceClient {
	return &replyServiceClient{cc}
}

func (c *replyServiceClient) AddReply(ctx context.Context, in *AddReplyReq, opts ...grpc.CallOption) (*AddReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_AddReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) DelReply(ctx context.Context, in *DelReplyReq, opts ...grpc.CallOption) (*DelReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DelReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_DelReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) LikeAction(ctx context.Context, in *LikeActionReq, opts ...grpc.CallOption) (*LikeActionRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LikeActionRes)
	err := c.cc.Invoke(ctx, ReplyService_LikeAction_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) DislikeAction(ctx context.Context, in *DislikeActionReq, opts ...grpc.CallOption) (*DislikeActionRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DislikeActionRes)
	err := c.cc.Invoke(ctx, ReplyService_DislikeAction_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) ReportReply(ctx context.Context, in *ReportReplyReq, opts ...grpc.CallOption) (*ReportReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReportReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_ReportReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) PinReply(ctx context.Context, in *PinReplyReq, opts ...grpc.CallOption) (*PinReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PinReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_PinReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) PageGetReply(ctx context.Context, in *PageGetReplyReq, opts ...grpc.CallOption) (*PageGetReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PageGetReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_PageGetReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) PageGetSubReply(ctx context.Context, in *PageGetSubReplyReq, opts ...grpc.CallOption) (*PageGetSubReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PageGetSubReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_PageGetSubReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) PageGetDetailedReply(ctx context.Context, in *PageGetReplyReq, opts ...grpc.CallOption) (*PageGetDetailedReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PageGetDetailedReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_PageGetDetailedReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) GetPinnedReply(ctx context.Context, in *GetPinnedReplyReq, opts ...grpc.CallOption) (*GetPinnedReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetPinnedReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_GetPinnedReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) CountReply(ctx context.Context, in *CountReplyReq, opts ...grpc.CallOption) (*CountReplyRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CountReplyRes)
	err := c.cc.Invoke(ctx, ReplyService_CountReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) BatchCountReply(ctx context.Context, in *BatchCountReplyRequest, opts ...grpc.CallOption) (*BatchCountReplyResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BatchCountReplyResponse)
	err := c.cc.Invoke(ctx, ReplyService_BatchCountReply_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) GetReplyLikeCount(ctx context.Context, in *GetReplyLikeCountReq, opts ...grpc.CallOption) (*GetReplyLikeCountRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetReplyLikeCountRes)
	err := c.cc.Invoke(ctx, ReplyService_GetReplyLikeCount_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) GetReplyDislikeCount(ctx context.Context, in *GetReplyDislikeCountReq, opts ...grpc.CallOption) (*GetReplyDislikeCountRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetReplyDislikeCountRes)
	err := c.cc.Invoke(ctx, ReplyService_GetReplyDislikeCount_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) CheckUserOnOjbect(ctx context.Context, in *CheckUserOnOjbectRequest, opts ...grpc.CallOption) (*CheckUserOnOjbectResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CheckUserOnOjbectResponse)
	err := c.cc.Invoke(ctx, ReplyService_CheckUserOnOjbect_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replyServiceClient) BatchCheckUserOnOjbect(ctx context.Context, in *BatchCheckUserOnOjbectRequest, opts ...grpc.CallOption) (*BatchCheckUserOnOjbectResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BatchCheckUserOnOjbectResponse)
	err := c.cc.Invoke(ctx, ReplyService_BatchCheckUserOnOjbect_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReplyServiceServer is the server API for ReplyService service.
// All implementations must embed UnimplementedReplyServiceServer
// for forward compatibility.
type ReplyServiceServer interface {
	// 发表评论
	AddReply(context.Context, *AddReplyReq) (*AddReplyRes, error)
	// 删除评论
	DelReply(context.Context, *DelReplyReq) (*DelReplyRes, error)
	// 赞
	LikeAction(context.Context, *LikeActionReq) (*LikeActionRes, error)
	// 踩
	DislikeAction(context.Context, *DislikeActionReq) (*DislikeActionRes, error)
	// 举报
	ReportReply(context.Context, *ReportReplyReq) (*ReportReplyRes, error)
	// 置顶评论
	PinReply(context.Context, *PinReplyReq) (*PinReplyRes, error)
	// 获取主评论信息
	PageGetReply(context.Context, *PageGetReplyReq) (*PageGetReplyRes, error)
	// 获取子评论信息
	PageGetSubReply(context.Context, *PageGetSubReplyReq) (*PageGetSubReplyRes, error)
	// 获取主评论详细信息
	PageGetDetailedReply(context.Context, *PageGetReplyReq) (*PageGetDetailedReplyRes, error)
	// 获取置顶评论
	GetPinnedReply(context.Context, *GetPinnedReplyReq) (*GetPinnedReplyRes, error)
	// 获取某个被评对象的评论数
	CountReply(context.Context, *CountReplyReq) (*CountReplyRes, error)
	// 获取多个被评论对象的评论数
	BatchCountReply(context.Context, *BatchCountReplyRequest) (*BatchCountReplyResponse, error)
	// 获取某条评论的点赞数
	GetReplyLikeCount(context.Context, *GetReplyLikeCountReq) (*GetReplyLikeCountRes, error)
	// 获取某条评论的点踩数
	GetReplyDislikeCount(context.Context, *GetReplyDislikeCountReq) (*GetReplyDislikeCountRes, error)
	// 获取某个用户是否评论了某个对象
	CheckUserOnOjbect(context.Context, *CheckUserOnOjbectRequest) (*CheckUserOnOjbectResponse, error)
	BatchCheckUserOnOjbect(context.Context, *BatchCheckUserOnOjbectRequest) (*BatchCheckUserOnOjbectResponse, error)
	mustEmbedUnimplementedReplyServiceServer()
}

// UnimplementedReplyServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedReplyServiceServer struct{}

func (UnimplementedReplyServiceServer) AddReply(context.Context, *AddReplyReq) (*AddReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddReply not implemented")
}
func (UnimplementedReplyServiceServer) DelReply(context.Context, *DelReplyReq) (*DelReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DelReply not implemented")
}
func (UnimplementedReplyServiceServer) LikeAction(context.Context, *LikeActionReq) (*LikeActionRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LikeAction not implemented")
}
func (UnimplementedReplyServiceServer) DislikeAction(context.Context, *DislikeActionReq) (*DislikeActionRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DislikeAction not implemented")
}
func (UnimplementedReplyServiceServer) ReportReply(context.Context, *ReportReplyReq) (*ReportReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportReply not implemented")
}
func (UnimplementedReplyServiceServer) PinReply(context.Context, *PinReplyReq) (*PinReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PinReply not implemented")
}
func (UnimplementedReplyServiceServer) PageGetReply(context.Context, *PageGetReplyReq) (*PageGetReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PageGetReply not implemented")
}
func (UnimplementedReplyServiceServer) PageGetSubReply(context.Context, *PageGetSubReplyReq) (*PageGetSubReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PageGetSubReply not implemented")
}
func (UnimplementedReplyServiceServer) PageGetDetailedReply(context.Context, *PageGetReplyReq) (*PageGetDetailedReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PageGetDetailedReply not implemented")
}
func (UnimplementedReplyServiceServer) GetPinnedReply(context.Context, *GetPinnedReplyReq) (*GetPinnedReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPinnedReply not implemented")
}
func (UnimplementedReplyServiceServer) CountReply(context.Context, *CountReplyReq) (*CountReplyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CountReply not implemented")
}
func (UnimplementedReplyServiceServer) BatchCountReply(context.Context, *BatchCountReplyRequest) (*BatchCountReplyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchCountReply not implemented")
}
func (UnimplementedReplyServiceServer) GetReplyLikeCount(context.Context, *GetReplyLikeCountReq) (*GetReplyLikeCountRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReplyLikeCount not implemented")
}
func (UnimplementedReplyServiceServer) GetReplyDislikeCount(context.Context, *GetReplyDislikeCountReq) (*GetReplyDislikeCountRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReplyDislikeCount not implemented")
}
func (UnimplementedReplyServiceServer) CheckUserOnOjbect(context.Context, *CheckUserOnOjbectRequest) (*CheckUserOnOjbectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckUserOnOjbect not implemented")
}
func (UnimplementedReplyServiceServer) BatchCheckUserOnOjbect(context.Context, *BatchCheckUserOnOjbectRequest) (*BatchCheckUserOnOjbectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchCheckUserOnOjbect not implemented")
}
func (UnimplementedReplyServiceServer) mustEmbedUnimplementedReplyServiceServer() {}
func (UnimplementedReplyServiceServer) testEmbeddedByValue()                      {}

// UnsafeReplyServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReplyServiceServer will
// result in compilation errors.
type UnsafeReplyServiceServer interface {
	mustEmbedUnimplementedReplyServiceServer()
}

func RegisterReplyServiceServer(s grpc.ServiceRegistrar, srv ReplyServiceServer) {
	// If the following call pancis, it indicates UnimplementedReplyServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ReplyService_ServiceDesc, srv)
}

func _ReplyService_AddReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).AddReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_AddReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).AddReply(ctx, req.(*AddReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_DelReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DelReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).DelReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_DelReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).DelReply(ctx, req.(*DelReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_LikeAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LikeActionReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).LikeAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_LikeAction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).LikeAction(ctx, req.(*LikeActionReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_DislikeAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DislikeActionReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).DislikeAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_DislikeAction_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).DislikeAction(ctx, req.(*DislikeActionReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_ReportReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).ReportReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_ReportReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).ReportReply(ctx, req.(*ReportReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_PinReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PinReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).PinReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_PinReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).PinReply(ctx, req.(*PinReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_PageGetReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PageGetReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).PageGetReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_PageGetReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).PageGetReply(ctx, req.(*PageGetReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_PageGetSubReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PageGetSubReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).PageGetSubReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_PageGetSubReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).PageGetSubReply(ctx, req.(*PageGetSubReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_PageGetDetailedReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PageGetReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).PageGetDetailedReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_PageGetDetailedReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).PageGetDetailedReply(ctx, req.(*PageGetReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_GetPinnedReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPinnedReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).GetPinnedReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_GetPinnedReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).GetPinnedReply(ctx, req.(*GetPinnedReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_CountReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CountReplyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).CountReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_CountReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).CountReply(ctx, req.(*CountReplyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_BatchCountReply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchCountReplyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).BatchCountReply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_BatchCountReply_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).BatchCountReply(ctx, req.(*BatchCountReplyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_GetReplyLikeCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReplyLikeCountReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).GetReplyLikeCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_GetReplyLikeCount_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).GetReplyLikeCount(ctx, req.(*GetReplyLikeCountReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_GetReplyDislikeCount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReplyDislikeCountReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).GetReplyDislikeCount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_GetReplyDislikeCount_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).GetReplyDislikeCount(ctx, req.(*GetReplyDislikeCountReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_CheckUserOnOjbect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckUserOnOjbectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).CheckUserOnOjbect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_CheckUserOnOjbect_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).CheckUserOnOjbect(ctx, req.(*CheckUserOnOjbectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReplyService_BatchCheckUserOnOjbect_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchCheckUserOnOjbectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplyServiceServer).BatchCheckUserOnOjbect(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReplyService_BatchCheckUserOnOjbect_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplyServiceServer).BatchCheckUserOnOjbect(ctx, req.(*BatchCheckUserOnOjbectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ReplyService_ServiceDesc is the grpc.ServiceDesc for ReplyService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ReplyService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "comment.sdk.v1.ReplyService",
	HandlerType: (*ReplyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddReply",
			Handler:    _ReplyService_AddReply_Handler,
		},
		{
			MethodName: "DelReply",
			Handler:    _ReplyService_DelReply_Handler,
		},
		{
			MethodName: "LikeAction",
			Handler:    _ReplyService_LikeAction_Handler,
		},
		{
			MethodName: "DislikeAction",
			Handler:    _ReplyService_DislikeAction_Handler,
		},
		{
			MethodName: "ReportReply",
			Handler:    _ReplyService_ReportReply_Handler,
		},
		{
			MethodName: "PinReply",
			Handler:    _ReplyService_PinReply_Handler,
		},
		{
			MethodName: "PageGetReply",
			Handler:    _ReplyService_PageGetReply_Handler,
		},
		{
			MethodName: "PageGetSubReply",
			Handler:    _ReplyService_PageGetSubReply_Handler,
		},
		{
			MethodName: "PageGetDetailedReply",
			Handler:    _ReplyService_PageGetDetailedReply_Handler,
		},
		{
			MethodName: "GetPinnedReply",
			Handler:    _ReplyService_GetPinnedReply_Handler,
		},
		{
			MethodName: "CountReply",
			Handler:    _ReplyService_CountReply_Handler,
		},
		{
			MethodName: "BatchCountReply",
			Handler:    _ReplyService_BatchCountReply_Handler,
		},
		{
			MethodName: "GetReplyLikeCount",
			Handler:    _ReplyService_GetReplyLikeCount_Handler,
		},
		{
			MethodName: "GetReplyDislikeCount",
			Handler:    _ReplyService_GetReplyDislikeCount_Handler,
		},
		{
			MethodName: "CheckUserOnOjbect",
			Handler:    _ReplyService_CheckUserOnOjbect_Handler,
		},
		{
			MethodName: "BatchCheckUserOnOjbect",
			Handler:    _ReplyService_BatchCheckUserOnOjbect_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "v1/reply.proto",
}
