package client

import (
	"fmt"

	grpc_fpl "github.com/go-fantasy/fpl/grpc"
	"google.golang.org/grpc"
)

//New creates a new FPL client object
func New() FPLClient {
	return &MyFPLClient{}
}

//Connect will connect the client to the gRPC endpoint
func (myFPLClient *MyFPLClient) Connect(port string) error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%v", port), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("error while connecting to gRPC server : %v", err)
	}

	client := grpc_fpl.NewFPLClient(conn)
	myFPLClient.conn = conn
	myFPLClient.client = client

	return nil
}

//GetClient returns the client object from the main struct
func (myFPLClient *MyFPLClient) GetClient() grpc_fpl.FPLClient {
	return myFPLClient.client
}

// Close handles closing gRPC connection
func (myFPLClient *MyFPLClient) Close() {
	if myFPLClient.conn != nil {
		myFPLClient.conn.Close()
	}
}
