package interceptor

import "github.com/zeromicro/go-zero/zrpc"

func InstallServerInterceptors(server *zrpc.RpcServer) {
	server.AddUnaryInterceptors(
		ServerErrorHandle,
		ServerMetadataExtract,
		ServerMetadateCheck(UidExistenceChecker),
		ServerValidateHandle,
	)
}
