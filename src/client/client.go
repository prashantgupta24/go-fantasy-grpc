package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/go-fantasy/src/server/grpc"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
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
	// c := pb.NewGreeterClient(conn)

	// // Contact the server and print out its response.
	// name := defaultName
	// if len(os.Args) > 1 {
	// 	name = os.Args[1]
	// }

	// r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	// if err != nil {
	// 	log.Fatalf("could not greet: %v", err)
	// }
	// log.Printf("Greeting: %s", r.Message)

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

	playerOccuranceDataStream, err := fplClient.GetDataForGameweek(ctx, &pb.GameweekReq{LeagueCode: 313, Gameweek: 9})

	for {
		playerOccuranceData, err := playerOccuranceDataStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", fplClient, err)
		}
		log.Printf("Player %v was selected by %v player/s!", playerOccuranceData.PlayerName, playerOccuranceData.PlayerOccuranceForGameweek)
	}
}
