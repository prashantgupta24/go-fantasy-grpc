package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	grpc_fpl "github.com/go-fantasy/fpl/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func (s *MyFPLServer) GetNumberOfPlayers(context.Context, *grpc_fpl.NumPlayerRequest) (*grpc_fpl.NumPlayers, error) {
	numPlayersInFPL := GetPlayerMapping(s)
	return &grpc_fpl.NumPlayers{NumPlayers: int64(numPlayersInFPL)}, nil
}

func (s *MyFPLServer) GetParticipantsInLeague(cxt context.Context, leagueCode *grpc_fpl.LeagueCode) (*grpc_fpl.NumParticipants, error) {
	numParticipants := GetParticipantsInLeague(s, int(leagueCode.LeagueCode))
	return &grpc_fpl.NumParticipants{NumParticipants: int64(numParticipants)}, nil
}

func (s *MyFPLServer) GetDataForGameweek(cxt context.Context, req *grpc_fpl.GameweekReq) (*grpc_fpl.PlayerOccuranceData, error) {
	GetPlayerMapping(s)
	GetParticipantsInLeague(s, int(req.LeagueCode))

	playerOccuranceForGameweek := make(map[string]int)
	fmt.Printf("Fetching data for gameweek %v\n", req.Gameweek)

	for _, participant := range s.leagueParticipants[0:10] {
		err := GetTeamInfoForParticipant(participant, int(req.Gameweek), playerOccuranceForGameweek, s)
		if err != nil {
			break
		}
	}
	if len(playerOccuranceForGameweek) > 0 {
		playerOccuranceData := &grpc_fpl.PlayerOccuranceData{
			PlayerOccurance: make(map[string]int32),
		}
		playerOccuranceResult := make(map[string]int32)
		for player, occurance := range playerOccuranceForGameweek {
			playerOccuranceResult[player] = int32(occurance)
		}
		playerOccuranceData.PlayerOccurance = playerOccuranceResult
		return playerOccuranceData, nil
	}
	return nil, nil
}

func (s *MyFPLServer) GetDataForAllGameweeks(req *grpc_fpl.LeagueCode, stream grpc_fpl.FPL_GetDataForAllGameweeksServer) error {
	GetPlayerMapping(s)
	GetParticipantsInLeague(s, int(req.LeagueCode))

	var wg sync.WaitGroup
	playerOccuranceChan := make(chan map[int]map[string]int)

	for gameweek := 1; gameweek <= gameweekMax; gameweek++ {
		wg.Add(1)
		go func(gameweek int) {
			playerOccuranceForGameweek := make(map[string]int)
			fmt.Printf("Fetching data for gameweek %v\n", gameweek)

			for _, participant := range s.leagueParticipants[0:10] {
				err := GetTeamInfoForParticipant(participant, gameweek, playerOccuranceForGameweek, s)
				if err != nil {
					break
				}
			}
			if len(playerOccuranceForGameweek) > 0 {
				playerOccuranceForGameweekMap := make(map[int]map[string]int)
				playerOccuranceForGameweekMap[gameweek] = playerOccuranceForGameweek
				playerOccuranceChan <- playerOccuranceForGameweekMap
			}
			wg.Done()
		}(gameweek)
	}

	go func() {
		wg.Wait()
		close(playerOccuranceChan)
	}()

	for playerOccuranceForGameweekMap := range playerOccuranceChan {
		for gameweekNum, playerOccuranceForGameweek := range playerOccuranceForGameweekMap {
			fmt.Printf("Data fetched for gameweek %v!\n", gameweekNum)
			s.playerOccurances[gameweekNum] = playerOccuranceForGameweek
		}
	}
	fileName := WriteToFile(s, int(req.LeagueCode))
	file, err := os.Open(fileName)
	if err != nil {
		return nil
	}
	defer func() {
		fmt.Println("removing temp file ", fileName)
		os.Remove(fileName)
	}()

	buf := make([]byte, 200)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		stream.Send(&grpc_fpl.AllGameweekData{
			Data: buf[:n],
		})
	}
}

//New is a helper function to create the main struct
func New() FPLServer {
	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	myFPLServer := &MyFPLServer{
		httpClient:       httpClient,
		playerMap:        make(map[int64]string),
		playerOccurances: make(map[int]map[string]int),
	}

	return myFPLServer
}

//Start will start the gRPC server
func (s *MyFPLServer) Start(port string) error {
	err := startgRPCServer(s, port)
	if err != nil {
		return err
	}
	return nil
}

//startgRPCServer is the official call to start the gRPC server
func startgRPCServer(myFPLServer *MyFPLServer, port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}
	// Creates a new gRPC server
	grpcServer := grpc.NewServer()
	grpc_fpl.RegisterFPLServer(grpcServer, myFPLServer)
	fmt.Println("started grpc server ...")
	grpcServer.Serve(lis)
	return nil
}
