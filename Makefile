# protoc:            ## Generate sources from protobuf definitions
# 	protoc -I . grpc/fpl.proto --go_out=plugins=grpc:server

mockgen-server:
	mockgen -source=fpl/server/types.go -destination=fpl/mock/mock_server.go

mockgen-client:
	mockgen github.com/go-fantasy/fpl/grpc FPLClient > fpl/mock-client/mock_client.go

vet:
	go vet $(shell glide nv)

server-start: vet
	go build -o ./example/bin/fplServer example/server/server_start.go && ./example/bin/fplServer

client-start:
	go build -o ./example/bin/fplServer example/client/client_main.go && ./example/bin/fplServer -l=313
test:
	go test fpl/server/*.go -v -failfast