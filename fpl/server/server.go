package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc/codes"

	grpc_fpl "github.com/go-fantasy/fpl/grpc"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

//GetNumberOfPlayers is the gRPC method to get number of players
func (s *MyFPLServer) GetNumberOfPlayers(context.Context, *grpc_fpl.NumPlayerRequest) (*grpc_fpl.NumPlayers, error) {
	numPlayersInFPL, err := GetPlayerMapping(s)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting player mapping : %v", err)
	}
	return &grpc_fpl.NumPlayers{NumPlayers: int64(numPlayersInFPL)}, nil
}

//GetParticipantsInLeague is the gRPC method to get number of participants in a league
func (s *MyFPLServer) GetParticipantsInLeague(cxt context.Context, leagueCode *grpc_fpl.LeagueCode) (*grpc_fpl.NumParticipants, error) {
	numParticipants, err := GetParticipantsInLeague(s, int(leagueCode.LeagueCode))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting participants in league : %v", err)
	}
	return &grpc_fpl.NumParticipants{NumParticipants: int64(numParticipants)}, nil
}

//GetDataForGameweek is the gRPC method to get player occurances for a single gameweek
func (s *MyFPLServer) GetDataForGameweek(cxt context.Context, req *grpc_fpl.GameweekReq) (*grpc_fpl.PlayerOccuranceData, error) {
	_, err := GetPlayerMapping(s)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting player mapping : %v", err)
	}
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

//GetDataForAllGameweeks is the gRPC method to get player occurances for all available gameweeks in a csv format
func (s *MyFPLServer) GetDataForAllGameweeks(req *grpc_fpl.LeagueCode, stream grpc_fpl.FPL_GetDataForAllGameweeksServer) error {
	_, err := GetPlayerMapping(s)
	if err != nil {
		return status.Errorf(codes.Internal, "error while getting player mapping : %v", err)
	}

	_, err = GetParticipantsInLeague(s, int(req.LeagueCode))
	if err != nil {
		return status.Errorf(codes.Internal, "error in GetParticipantsInLeague : %v", err)
	}

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
	fileName, err := WriteToFile(s, int(req.LeagueCode))
	if err != nil {
		return status.Errorf(codes.Internal, "error while writing to file %v : %v", fileName, err)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return status.Errorf(codes.Internal, "error while opening file %v : %v", fileName, err)
	}

	defer func() error {
		fmt.Println("removing temp file ", fileName)
		err := os.Remove(fileName)
		if err != nil {
			return status.Errorf(codes.Internal, "error while deleting temp file %v : %v", fileName, err)
		}
		return nil
	}()

	buf := make([]byte, 200)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Internal, "error while writing to file %v : %v", fileName, err)
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

//MakeRequest is used for making http requests
func (s *MyFPLServer) MakeRequest(URL string) ([]byte, error) {

	var err error
	customErr := errors.Errorf("error with request to %v : %v", URL, err)

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, customErr
	}

	req.Header.Set("User-Agent", "pg-fpl")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, customErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, customErr
	}

	return body, nil
}

//GetPlayerOccurances gets the player occurances
func (s *MyFPLServer) GetPlayerOccurances() map[int]map[string]int {
	return s.playerOccurances
}

//GetPlayerMap gets the player map created for FPL
func (s *MyFPLServer) GetPlayerMap() map[int64]string {
	return s.playerMap
}

//startgRPCServer is the official call to start the gRPC server
func startgRPCServer(myFPLServer *MyFPLServer, port string) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return status.Errorf(codes.Internal, "error while starting server : %v", err)
	}
	// Creates a new gRPC server
	grpcServer := grpc.NewServer()
	grpc_fpl.RegisterFPLServer(grpcServer, myFPLServer)
	fmt.Printf("started grpc server at port %v ...\n", port)
	grpcServer.Serve(lis)
	return nil
}
