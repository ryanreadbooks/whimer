
.PHONY: install-dev
install-dev:
	@sh -c ./install.sh

IMAGE_VER := latest

.PHONY: build-image
build-image:
	docker buildx build -t s1-gate:${IMAGE_VER} .
