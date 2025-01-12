.PHONY: buf
buf:
	buf generate

.PHONY: dev-run
dev-run:
	cd .bootstrap/docker && make up
	goreman start