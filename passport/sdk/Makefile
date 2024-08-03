PROTO_PATH := $(shell echo $$PROTO_PATH)

.PHONY: access
access:
	protoc -I ${PROTO_PATH} -I user -I access access.proto \
		--go_out=access \
		--go_opt=module=github.com/ryanreadbooks/whimer/passport/sdk/access \
		--go-grpc_out=access \
		--go-grpc_opt=module=github.com/ryanreadbooks/whimer/passport/sdk/access

.PHONY: user
user:
	protoc -I ${PROTO_PATH} -I user user.proto \
		--go_out=user \
		--go_opt=module=github.com/ryanreadbooks/whimer/passport/sdk/user \
		--go-grpc_out=user \
		--go-grpc_opt=module=github.com/ryanreadbooks/whimer/passport/sdk/user

.PHONY: proto
proto: user access
