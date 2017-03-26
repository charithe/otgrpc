ROOT_PKG:=github.com/charithe/otgrpc
PROTO_PKG:=test

.PHONY: test

$(PROTO_PKG)/test.pb.go:
	protoeasy --gogo --go-import-path=$(ROOT_PKG) --grpc $(PROTO_PKG)/

test:
	go test -v ./...
	go tool vet .


