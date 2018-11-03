package main

import (
	"log"

	flag "github.com/spf13/pflag"

	"github.com/go-fantasy/fpl/server"
	"github.com/spf13/viper"
)

func main() {
	flag.StringP("port", "p", "50051", "Port for the gRPC server")
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	myFPLServer := server.New()
	err := myFPLServer.Start(viper.GetString("port"))
	if err != nil {
		log.Fatalf("Error starting gRPC server! %v", err)
	}
}
