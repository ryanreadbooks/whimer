.PHONY: buf
buf:
	buf generate

.PHONY: dev-run
dev-run:
	cd .bootstrap/docker && make up
	goreman start

.PHONY: tidy-mods
tidy-mods:
	@echo "Running go mod tidy in top-level modules..."
	@set -e; \
	find . -mindepth 1 -maxdepth 1 -type d | while read -r dir; do \
		if [ -f "$$dir/go.mod" ]; then \
			printf "==> tidy in %s\n" "$$dir"; \
			(cd "$$dir" && go mod tidy); \
		fi; \
	done