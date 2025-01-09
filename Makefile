.PHONY: build
build:
	mkdir -p out
	go build -o out/govc-exporter -gcflags CGO_ENABLED=0 ./cmd/exporter 
	go build -o out/vcenter-object-exporter -gcflags CGO_ENABLED=0 ./cmd/vcenter-object-exporter

goreleaser-build:
	goreleaser build --clean --snapshot