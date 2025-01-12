PROTO_PATH := $(shell echo $$PROTO_PATH)

.PHONY: sdk-v1
sdk-v1:
	protoc -I ${PROTO_PATH} -I v1 counter.proto \
		--go_out=v1 \
		--go_opt=module=github.com/ryanreadbooks/whimer/counter/sdk/v1 \
		--go-grpc_out=v1 \
		--go-grpc_opt=module=github.com/ryanreadbooks/whimer/counter/sdk/v1