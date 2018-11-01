package main

import (
	"context"
	"log"
	"time"

	pb "github.com/go-fantasy/fpl/grpc"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	fplClient := pb.NewFPLClient(conn)

	numPlayers, err := fplClient.GetNumberOfPlayers(ctx, &pb.NumPlayerRequest{})

	if err != nil {
		log.Fatalf("could not fetch: %v", err)
	}

	log.Printf("There are %v players in fpl!", numPlayers.NumPlayers)

	leagueCode := int64(313)
	numParticipants, err := fplClient.GetParticipantsInLeague(ctx, &pb.LeagueCode{LeagueCode: leagueCode})

	if err != nil {
		log.Fatalf("could not fetch: %v", err)
	}

	log.Printf("There are %v participants in %v league!", numParticipants.NumParticipants, leagueCode)

	playerOccurance, err := fplClient.GetDataForGameweek(ctx, &pb.GameweekReq{LeagueCode: 313, Gameweek: 9})

	// for {
	// 	playerOccuranceData, err := playerOccuranceDataStream.Recv()
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatalf("%v.ListFeatures(_) = _, %v", fplClient, err)
	// 	}
	// 	log.Printf("Player %v was selected by %v player/s!", playerOccuranceData.PlayerName, playerOccuranceData.PlayerOccuranceForGameweek)
	// }

	for player, occurance := range playerOccurance.PlayerOccurance {
		log.Printf("Player %v was selected by %v player/s!", player, occurance)
	}

	// csvFile, err := fplClient.GetDataForAllGameweeks(ctx, &pb.LeagueCode{LeagueCode: leagueCode})

	// f, err := os.Create("dataFile")
	// if err != nil {
	// 	log.Fatal("unable to create file")
	// }
	// defer f.Close()
	// for {
	// 	buf, err := csvFile.Recv()
	// 	if err != nil {
	// 		break
	// 	}
	// 	n2, err := f.Write()
	// 	if err != nil {
	// 		break
	// 	}
	// }
}
