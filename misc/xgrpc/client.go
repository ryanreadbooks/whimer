package xgrpc

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 快速创建带通用拦截器的grpc客户端连接
func NewClientFromDiscovery(conf xconf.Discovery) (zrpc.Client, error) {
	cli, err := zrpc.NewClient(
		conf.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject),
	)

	return cli, err
}

// 泛型版本
// 快速创建带通用来拦截器的grpc客户端连接
func NewClient[T any, P func(cc grpc.ClientConnInterface) T](
	conf xconf.Discovery, constructor P,
) (ret T, err error) {
	cli, err := zrpc.NewClient(
		conf.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientErrorHandler),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject),
	)
	if err != nil {
		return
	}

	return constructor(cli.Conn()), nil
}

func MustNewClient[T any, P func(cc grpc.ClientConnInterface) T](
	conf xconf.Discovery, constructor P,
) (ret T) {
	c, err := NewClient(conf, constructor)
	if err != nil {
		panic(err)
	}

	return c
}

// 创建带通用拦截器的grpc客户端连接
func NewClientConn(conf xconf.Discovery) (*grpc.ClientConn, error) {
	cli, err := zrpc.NewClient(
		conf.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientErrorHandler),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject),
		zrpc.WithDialOption(grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: 8 * time.Second,
		})),
	)
	if err != nil {
		return nil, err
	}

	return cli.Conn(), nil
}

func NewClientConnWithoutInterceptors(conf xconf.Discovery) (*grpc.ClientConn, error) {
	cli, err := zrpc.NewClient(
		conf.AsZrpcClientConf(),
		zrpc.WithDialOption(grpc.WithConnectParams(grpc.ConnectParams{
			Backoff:           backoff.DefaultConfig,
			MinConnectTimeout: 8 * time.Second,
		})),
	)
	if err != nil {
		return nil, err
	}

	return cli.Conn(), nil
}

var _ grpc.ClientConnInterface = &UnreadyClientConn{}

type UnreadyClientConn struct{}

func NewUnreadyClientConn() *UnreadyClientConn {
	return &UnreadyClientConn{}
}

func (*UnreadyClientConn) Invoke(ctx context.Context,
	method string,
	args any,
	reply any,
	opts ...grpc.CallOption,
) error {
	return status.Error(codes.FailedPrecondition, xerror.ErrDepNotReady.Error())
}

func (*UnreadyClientConn) NewStream(ctx context.Context,
	desc *grpc.StreamDesc,
	method string,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return nil, status.Error(codes.FailedPrecondition, xerror.ErrDepNotReady.Error())
}

// 重新连接
func RetryConnectConn(c xconf.Discovery, connector func(cc grpc.ClientConnInterface)) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// try to reconnect
		logx.Infof("retrying to connect to conn at %s(%v)", c.Key, c.Hosts)
		cc, err := NewClientConn(c)
		if err != nil {
			logx.Infof("retrying to connect to conn at %s(%v) failed again: %v", c.Key, c.Hosts, err)
			continue
		}

		// retry connect succeeded
		// ignore concurrency
		connector(cc)
		logx.Infof("retry connecting %s(%v) loop succeeded", c.Key, c.Hosts)
		break
	}
}

func RetryConnectConnInBackground(c xconf.Discovery, connect func(cc grpc.ClientConnInterface)) {
	concurrent.SafeGo(func() {
		RetryConnectConn(c, connect)
	})
}

// Deprecated: use NewRecoverableClientConn instead
func NewRecoverableClient[T any](c xconf.Discovery,
	getter func(grpc.ClientConnInterface) T,
	fallback func(T),
	options ...ClientOption,
) T {
	option := defaultClientOption()
	for _, o := range options {
		o(option)
	}

	var (
		cc  grpc.ClientConnInterface
		err error
	)

	if option.withoutDefaultInterceptor {
		cc, err = NewClientConnWithoutInterceptors(c)
	} else {
		cc, err = NewClientConn(c)
	}

	if err != nil {
		logx.Infof("can not init grpc client %v", err)
		unready := getter(NewUnreadyClientConn())
		RetryConnectConnInBackground(c, func(cc grpc.ClientConnInterface) {
			// we ignore concurrent problem here
			newT := getter(cc)
			fallback(newT)
		})
		return unready
	}

	return getter(cc)
}

type RecoverableClientConn struct {
	ready atomic.Bool

	unreadyConn grpc.ClientConnInterface
	readyConn   atomic.Value // grpc.ClientConnInterface
}

var _ grpc.ClientConnInterface = &RecoverableClientConn{}

func (c *RecoverableClientConn) Invoke(ctx context.Context,
	method string,
	args any,
	reply any,
	opts ...grpc.CallOption,
) error {
	if val := c.readyConn.Load(); val != nil && c.ready.Load() {
		if cci := val.(grpc.ClientConnInterface); cci != nil {
			return cci.Invoke(ctx, method, args, reply, opts...)
		}
	}

	return c.unreadyConn.Invoke(ctx, method, args, reply, opts...)
}

func (c *RecoverableClientConn) NewStream(ctx context.Context,
	desc *grpc.StreamDesc,
	method string,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	if val := c.readyConn.Load(); val != nil && c.ready.Load() {
		if cci := val.(grpc.ClientConnInterface); cci != nil {
			return cci.NewStream(ctx, desc, method, opts...)
		}
	}

	return c.unreadyConn.NewStream(ctx, desc, method, opts...)
}

func NewRecoverableClientConn(c xconf.Discovery, options ...ClientOption) grpc.ClientConnInterface {
	option := defaultClientOption()
	for _, o := range options {
		o(option)
	}

	var (
		cc  grpc.ClientConnInterface
		err error
	)

	if option.withoutDefaultInterceptor {
		cc, err = NewClientConnWithoutInterceptors(c)
	} else {
		cc, err = NewClientConn(c)
	}

	if err == nil {
		return cc
	}

	// 首次连接失败 返回一个后台重试的client conn
	rCc := &RecoverableClientConn{
		unreadyConn: NewUnreadyClientConn(),
		readyConn:   atomic.Value{},
	}

	concurrent.SafeGo(func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			logx.Infof("retrying-v2 to connect to conn at %s(%v)", c.Key, c.Hosts)
			var (
				newCc grpc.ClientConnInterface
				err   error
			)
			if option.withoutDefaultInterceptor {
				newCc, err = NewClientConnWithoutInterceptors(c)
			} else {
				newCc, err = NewClientConn(c)
			}
			if err != nil {
				logx.Infof("retrying-v2 to connect to conn at %s(%v) failed again: %v", c.Key, c.Hosts, err)
				continue
			}

			logx.Infof("retry-v2 connecting %s(%v) loop succeeded", c.Key, c.Hosts)
			// set to ready
			rCc.readyConn.Store(newCc)
			rCc.ready.Store(true)
			break
		}
	})

	return rCc
}
