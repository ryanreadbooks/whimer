.PHONY: buf
buf:
	buf generate

.PHONY: dev-run
dev-run:
	goreman start