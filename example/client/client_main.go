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
		fmt.Println("league code is required")
		os.Exit(1)
	}
	flag.Int64P("league", "l", 313, "League code")
	flag.StringP("port", "p", "50051", "Port to connect to the gRPC server")
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)
}

func main() {
	parseFlags()
	myFPLClientObj := client.New()
	err := myFPLClientObj.Connect(viper.GetString("port"))
	if err != nil {
		log.Fatal("error creating client")
	}
	defer myFPLClientObj.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	myFPLClient := myFPLClientObj.GetClient()
	leagueCode := viper.GetInt64("league")

	// //First gRPC method
	// numPlayers, err := fplClient.GetNumberOfPlayers(ctx, &grpc_fpl.NumPlayerRequest{})
	// if err != nil {
	// 	log.Fatalf("could not fetch: %v", err)
	// }
	// log.Printf("There are %v players in fpl!", numPlayers.NumPlayers)

	// //Second method
	// numParticipants, err := fplClient.GetParticipantsInLeague(ctx, &grpc_fpl.LeagueCode{LeagueCode: leagueCode})
	// if err != nil {
	// 	log.Fatalf("could not fetch: %v", err)
	// }
	// log.Printf("There are %v participants in %v league!", numParticipants.NumParticipants, leagueCode)

	// //Third method
	// playerOccurance, err := fplClient.GetDataForGameweek(ctx, &grpc_fpl.GameweekReq{LeagueCode: 313, Gameweek: 9})
	// for player, occurance := range playerOccurance.PlayerOccurance {
	// 	log.Printf("Player %v was selected by %v player/s!", player, occurance)
	// }

	//Fourth method
	resultFile, err := os.Create(fmt.Sprintf("../data/dataFile-%v-%v.csv", time.Now().Format("2006-01-02"), leagueCode))
	if err != nil {
		log.Fatal("Unable to create file : ", err)
	}
	defer func() {
		resultFile.Sync()
		resultFile.Close()
	}()

	csvFile, err := myFPLClient.GetDataForAllGameweeks(ctx, &grpc_fpl.LeagueCode{LeagueCode: leagueCode})
	if err != nil {
		log.Fatal("Unable to fetch data from gRPC method : ", err)
	}
	for {
		buf, err := csvFile.Recv()
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
}
