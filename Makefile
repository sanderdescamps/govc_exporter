.PHONY: build

build:
	goreleaser build --clean --snapshot

build-simple:
	go mod tidy
	mkdir -p out
	go build -o out/govc-exporter -gcflags CGO_ENABLED=0 ./cmd/exporter 
	go build -o out/vcenter-object-exporter -gcflags CGO_ENABLED=0 ./cmd/vcenter-object-exporter

vendor:
	go mod vendor

test-vcenter:
	./scripts/gi-start-vcsim.sh