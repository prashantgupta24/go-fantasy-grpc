package client

import (
	grpc_fpl "github.com/go-fantasy/fpl/grpc"
	"google.golang.org/grpc"
)

//FPLClient is the main client interface to talk to the server
type FPLClient interface {
	Connect(string) error
	GetClient() grpc_fpl.FPLClient
	Close()
}

//MyFPLClient is my implementation of the official client
type MyFPLClient struct {
	conn   *grpc.ClientConn
	client grpc_fpl.FPLClient
}
