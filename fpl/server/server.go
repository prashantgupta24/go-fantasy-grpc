package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc/codes"

	grpc_fpl "github.com/go-fantasy/fpl/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

//GetNumberOfPlayers is the gRPC method to get number of players
func (s *MyFPLServer) GetNumberOfPlayers(context.Context, *grpc_fpl.NumPlayerRequest) (*grpc_fpl.NumPlayers, error) {
	playerMap, err := s.Scraper.GetPlayerMapping()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting player mapping : %v", err)
	}
	numPlayersInFPL := len(playerMap)
	return &grpc_fpl.NumPlayers{NumPlayers: int64(numPlayersInFPL)}, nil
}

//GetParticipantsInLeague is the gRPC method to get number of participants in a league
func (s *MyFPLServer) GetParticipantsInLeague(cxt context.Context, leagueCode *grpc_fpl.LeagueCode) (*grpc_fpl.NumParticipants, error) {
	leagueParticipants, err := s.Scraper.GetParticipantsInLeague(int(leagueCode.LeagueCode))
	numParticipants := len(*leagueParticipants)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting participants in league : %v", err)
	}
	return &grpc_fpl.NumParticipants{NumParticipants: int64(numParticipants)}, nil
}

//GetDataForGameweek is the gRPC method to get player occurances for a single gameweek
func (s *MyFPLServer) GetDataForGameweek(cxt context.Context, req *grpc_fpl.GameweekReq) (*grpc_fpl.PlayerOccuranceData, error) {
	playerMap, err := s.Scraper.GetPlayerMapping()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting player mapping : %v", err)
	}
	s.PlayerMap = playerMap

	participants, err := s.Scraper.GetParticipantsInLeague(int(req.LeagueCode))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while getting participants in league : %v", err)
	}
	s.LeagueParticipants = participants

	leagueParticipants := *s.LeagueParticipants
	var topLeagueParticipants []int64
	if len(leagueParticipants) > 10 {
		topLeagueParticipants = leagueParticipants[0:10]
	} else {
		topLeagueParticipants = leagueParticipants[:]
	}
	fmt.Printf("Fetching data for gameweek %v\n", req.Gameweek)
	playerOccuranceForGameweek, err := s.Scraper.GetTeamInfoForParticipant(playerMap, int(req.Gameweek), &topLeagueParticipants)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error while Fetching data for gameweek %v : %v", int(req.Gameweek), err)
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
	playerMap, err := s.Scraper.GetPlayerMapping()
	if err != nil {
		return status.Errorf(codes.Internal, "error while getting player mapping : %v", err)
	}
	s.PlayerMap = playerMap

	participants, err := s.Scraper.GetParticipantsInLeague(int(req.LeagueCode))
	if err != nil {
		return status.Errorf(codes.Internal, "error in GetParticipantsInLeague : %v", err)
	}
	s.LeagueParticipants = participants

	var wg sync.WaitGroup
	playerOccuranceChan := make(chan map[int]map[string]int)

	for gameweek := 1; gameweek <= GameweekMax; gameweek++ {
		wg.Add(1)
		go func(gameweek int, playerOccuranceChan chan map[int]map[string]int) {
			defer wg.Done()

			leagueParticipants := *s.LeagueParticipants
			var topLeagueParticipants []int64
			if len(leagueParticipants) > 10 {
				topLeagueParticipants = leagueParticipants[0:10]
			} else {
				topLeagueParticipants = leagueParticipants[:]
			}
			fmt.Printf("Fetching data for gameweek %v\n", gameweek)

			playerOccuranceForGameweek, err := s.Scraper.GetTeamInfoForParticipant(playerMap, gameweek, &topLeagueParticipants)
			if err != nil {
				//return nil, status.Errorf(codes.Internal, "error while Fetching data for gameweek %v : %v", int(req.Gameweek), err)
			}
			if len(playerOccuranceForGameweek) > 0 {
				playerOccuranceForGameweekMap := make(map[int]map[string]int)
				playerOccuranceForGameweekMap[gameweek] = playerOccuranceForGameweek
				playerOccuranceChan <- playerOccuranceForGameweekMap
			}
		}(gameweek, playerOccuranceChan)
	}

	go func() {
		wg.Wait()
		close(playerOccuranceChan)
	}()

	for playerOccuranceForGameweekMap := range playerOccuranceChan {
		for gameweekNum, playerOccuranceForGameweek := range playerOccuranceForGameweekMap {
			fmt.Printf("Data fetched for gameweek %v!\n", gameweekNum)
			s.PlayerOccurances[gameweekNum] = playerOccuranceForGameweek
		}
	}
	fileName, err := s.Scraper.WriteToFile(s.PlayerOccurances, int(req.LeagueCode))
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("error while writing to file %v : %v", fileName, err))
	}
	defer func() error {
		fmt.Println("removing temp file ", fileName)
		err := os.Remove(fileName)
		if err != nil {
			return status.Errorf(codes.Internal, "error while deleting temp file %v : %v", fileName, err)
		}
		return nil
	}()

	file, err := os.Open(fileName)
	if err != nil {
		return status.Errorf(codes.Internal, "error while opening file %v : %v", fileName, err)
	}

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
		PlayerMap:        make(map[int64]string),
		PlayerOccurances: make(map[int]map[string]int),
		Scraper: &MyFPLScraper{
			Client: &MyFPLClient{
				HttpClient: httpClient,
			},
		},
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
		return status.Errorf(codes.Internal, "error while starting server : %v", err)
	}
	// Creates a new gRPC server
	grpcServer := grpc.NewServer()
	grpc_fpl.RegisterFPLServer(grpcServer, myFPLServer)
	fmt.Printf("started grpc server at port %v ...\n", port)
	grpcServer.Serve(lis)
	return nil
}
