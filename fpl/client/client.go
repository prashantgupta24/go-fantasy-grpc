package client

import (
	"fmt"

	grpc_fpl "github.com/go-fantasy/fpl/grpc"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

//New creates a new FPL client object
func New() (grpc_fpl.FPLClient, func(), error) {
	port := viper.GetString("port")
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%v", port), grpc.WithInsecure())
	if err != nil {
		return nil, nil, fmt.Errorf("error while connecting to gRPC server at port %v: %v", port, err)
	}

	client := grpc_fpl.NewFPLClient(conn)

	cleanup := func() {
		if conn != nil {
			conn.Close()
		}
	}
	return client, cleanup, nil
}
