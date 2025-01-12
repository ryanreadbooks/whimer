PROTO_PATH := $(shell echo $$PROTO_PATH)

.PHONY: reply
reply:
	protoc -I ${PROTO_PATH} -I . reply.proto \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		--go-http_out=. \
		--go-http_opt=paths=source_relative \