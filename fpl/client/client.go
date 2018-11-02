package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/go-fantasy/fpl/grpc"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

//FplClient is the official struct
type FplClient struct {
	conn   *grpc.ClientConn
	client pb.FPLClient
}

//CreateNewFPLClient creates a new FPL client
func CreateNewFPLClient() (*FplClient, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := pb.NewFPLClient(conn)

	return &FplClient{
		conn:   conn,
		client: client,
	}, nil
}

func (fplClient *FplClient) close() {
	if fplClient.conn != nil {
		fplClient.conn.Close()
	}
}

func main() {

	fplClientObj, err := CreateNewFPLClient()
	if err != nil {
		log.Fatal("error creating client")
	}

	fplClient := fplClientObj.client
	defer fplClientObj.close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	leagueCode := int64(313)

	// //First gRPC method
	// numPlayers, err := fplClient.GetNumberOfPlayers(ctx, &pb.NumPlayerRequest{})
	// if err != nil {
	// 	log.Fatalf("could not fetch: %v", err)
	// }

	// log.Printf("There are %v players in fpl!", numPlayers.NumPlayers)

	// //Second method
	// numParticipants, err := fplClient.GetParticipantsInLeague(ctx, &pb.LeagueCode{LeagueCode: leagueCode})
	// if err != nil {
	// 	log.Fatalf("could not fetch: %v", err)
	// }

	// log.Printf("There are %v participants in %v league!", numParticipants.NumParticipants, leagueCode)

	// //Third method
	// playerOccurance, err := fplClient.GetDataForGameweek(ctx, &pb.GameweekReq{LeagueCode: 313, Gameweek: 9})

	// for player, occurance := range playerOccurance.PlayerOccurance {
	// 	log.Printf("Player %v was selected by %v player/s!", player, occurance)
	// }

	//Fourth method
	resultFile, err := os.Create(fmt.Sprintf("data/dataFile-%v-%v.csv", time.Now().Format("2006-01-02"), leagueCode))
	if err != nil {
		log.Fatal("unable to create file")
	}
	defer func() {
		resultFile.Sync()
		resultFile.Close()
	}()

	csvFile, err := fplClient.GetDataForAllGameweeks(ctx, &pb.LeagueCode{LeagueCode: leagueCode})
	for {
		buf, err := csvFile.Recv()
		if err == io.EOF {
			log.Println("File transfer complete!")
			break
		}
		if err != nil {
			log.Fatal("Error while reading from file", err)
		}
		_, err = resultFile.Write(buf.Data)
		if err != nil {
			break
		}
	}
}
