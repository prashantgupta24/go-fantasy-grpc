package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	grpc_fpl "github.com/go-fantasy/fpl/grpc"

	//pflag is a drop-in replacement of Go's native flag package
	"github.com/go-fantasy/fpl/client"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func parseFlags() {
	if len(os.Args) < 2 {
		fmt.Println("league code is required, usage is -l <league code>")
		os.Exit(1)
	}
	flag.Int64P("league", "l", 313, "League code")
	flag.Int64P("gameweek", "g", 1, "Gameweek")
	flag.StringP("port", "p", "50051", "Port to connect to the gRPC server")
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)
}

func main() {
	parseFlags()
	grpcClient, cleanup, err := client.New()
	if err != nil {
		log.Fatalf("error creating client %v ", err)
	}
	defer cleanup()

	leagueCode := viper.GetInt64("league")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	//First gRPC method
	//getNumPlayers(ctx, grpcClient)

	//Second method
	//getParticipantsInLeague(ctx, grpcClient, leagueCode)

	// //Third method
	//gameweek := viper.GetInt64("gameweek")
	//getDataForGameweek(ctx, grpcClient, gameweek)

	//Fourth method
	file := getDataForAllGameweeks(ctx, grpcClient, leagueCode)
	fmt.Printf("Received %v file from server! \n", file.Name())
}

func getNumPlayers(ctx context.Context, grpcClient grpc_fpl.FPLClient) {
	numPlayers, err := grpcClient.GetNumberOfPlayers(ctx, &grpc_fpl.NumPlayerRequest{})
	if err != nil {
		log.Fatalf("could not fetch GetNumberOfPlayers: %v", err)
	}
	log.Printf("There are %v players in fpl!", numPlayers.NumPlayers)
}

func getParticipantsInLeague(ctx context.Context, grpcClient grpc_fpl.FPLClient, leagueCode int64) {
	numParticipants, err := grpcClient.GetParticipantsInLeague(ctx, &grpc_fpl.LeagueCode{LeagueCode: leagueCode})
	if err != nil {
		log.Fatalf("could not fetch GetParticipantsInLeague: %v", err)
	}
	log.Printf("There are %v participants in league %v!", numParticipants.NumParticipants, leagueCode)
}

func getDataForGameweek(ctx context.Context, grpcClient grpc_fpl.FPLClient, gameweek int64) {
	playerOccurance, err := grpcClient.GetDataForGameweek(ctx, &grpc_fpl.GameweekReq{LeagueCode: 313, Gameweek: gameweek})
	if err != nil {
		log.Fatalf("could not fetch GetDataForGameweek: %v", err)
	}
	for player, occurance := range playerOccurance.PlayerOccurance {
		log.Printf("Player %v was selected by \t\t%v player/s!", player, occurance)
	}
}

func getDataForAllGameweeks(ctx context.Context, grpcClient grpc_fpl.FPLClient, leagueCode int64) *os.File {
	resultFile, err := os.Create(fmt.Sprintf("example/data/dataFile-%v-%v.csv", time.Now().Format("2006-01-02-15-04"), leagueCode))
	if err != nil {
		log.Fatal("Unable to create file : ", err)
	}
	defer func() {
		resultFile.Sync()
		resultFile.Close()
	}()

	stream, err := grpcClient.GetDataForAllGameweeks(ctx, &grpc_fpl.LeagueCode{LeagueCode: leagueCode})
	if err != nil {
		log.Fatal("Unable to fetch data for GetDataForAllGameweeks gRPC method : ", err)
	}
	for {
		buf, err := stream.Recv()
		if err == io.EOF {
			log.Println("File transfer complete!")
			break
		}
		if err != nil {
			log.Fatal("Error while reading from file: ", err)
		}
		_, err = resultFile.Write(buf.Data)
		if err != nil {
			break
		}
	}
	return resultFile
}
