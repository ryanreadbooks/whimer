package testproto

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/apiextension/protobuf/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type serverImpl struct {
	UnimplementedExtensionTestServiceServer
}

func (s *serverImpl) GetTest(ctx context.Context, in *GetTestRequest) (*GetTestResponse, error) {
	return &GetTestResponse{}, nil
}

func (s *serverImpl) MustGet(ctx context.Context, in *MustGetRequest) (*MustGetResponse, error) {
	return &MustGetResponse{}, nil
}

func TestMethodOption(t *testing.T) {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				t.Logf("server interceptor = %s\n", info.FullMethod)
				methodName := strings.ReplaceAll(info.FullMethod[1:], "/", ".")
				desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(methodName))
				t.Log(err)
				t.Log(desc)
				methodDesc, ok := desc.(protoreflect.MethodDescriptor)
				if ok {
					t.Log(methodDesc.FullName())
					// we can get options here
					option := methodDesc.Options()
					if option != nil {
						ext := proto.GetExtension(option, options.E_Method)
						if ext != nil {
							methodOption, ok := ext.(*options.WhimerMethodOption)
							if ok && methodOption != nil {
								t.Logf("method %s skip option = %v\n", info.FullMethod, methodOption.SkipMetadataUidCheck)
							} else {
								t.Logf("method %s skip option is empty\n", info.FullMethod)
							}
						} else {
							t.Logf("method %s has no ext\n", info.FullMethod)
						}
					} else {
						t.Logf("method %s has no option\n", info.FullMethod)
					}
				}

				// get service name
				serviceName := protoreflect.FullName(methodName).Parent()
				serviceDesc,err  := protoregistry.GlobalFiles.FindDescriptorByName(serviceName)
				if err == nil {
					serviceOption := serviceDesc.Options()
					serviceExt := proto.GetExtension(serviceOption, options.E_Service)
					whimerOption, ok := serviceExt.(*options.WhimerServiceOption) 
					if ok && whimerOption != nil {
						t.Logf("service %s got option = %v", serviceName, whimerOption)
					}
				}

				return handler(ctx, req)
			},
		),
	)

	addr := "127.0.0.1:60000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	RegisterExtensionTestServiceServer(server, &serverImpl{})
	go func() {
		err = server.Serve(lis)
		if err != nil {
			t.Error(err)
		}
	}()

	defer server.GracefulStop()
	time.Sleep(time.Millisecond * 20)
	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	cli := NewExtensionTestServiceClient(cc)
	_, err = cli.GetTest(t.Context(), &GetTestRequest{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = cli.MustGet(t.Context(), &MustGetRequest{})
	if err != nil {
		t.Fatal(err)
	}
}
