package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	pb "github.com/go-fantasy/src/server/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port            = ":50051"
	teamURL         = "https://fantasy.premierleague.com/drf/entry/%v/event/%v/picks"
	allPlayersURL   = "https://fantasy.premierleague.com/drf/bootstrap-static"
	participantsURL = "https://fantasy.premierleague.com/drf/leagues-classic-standings/%v?phase=1&le-page=1&ls-page=1"
	csvFileName     = "data/result-%v-%v.csv"
)

type fantasyMain struct {
	httpClient         *http.Client
	playerMap          map[int64]string
	leagueParticipants []int64
	playerOccurances   map[int]map[string]int
}

/* Structure of JSON

picks
    0
    element	260
    1
    element	247
*/
type ParticipantTeamInfo struct {
	TeamPlayers []TeamPlayers `json:"picks"`
}
type TeamPlayers struct {
	Element int64 `json:"element"`
}

/* Structure of JSON

elements
    0
    id	1
    photo	"11334.jpg"
    web_name	"Cech"
    team_code	3
    status	"i"
    code	11334
    first_name	"Petr"
    second_name	"Cech"
    squad_number	1

    1
    id	2
    photo	"80201.jpg"
    web_name	"Leno"
    team_code	3
    status	"a"
    code	80201
    first_name	"Bernd"
    second_name	"Leno"
    squad_number	19
*/
type AllPlayers struct {
	Players []Players `json:"elements"`
}
type Players struct {
	ID      int64  `json:"id"`
	WebName string `json:"web_name"`
}

/* Structure of JSON

standings
    has_next	true
    number	1
    results
        0
        id	13987896
        rank	1
        last_rank	1
        rank_sort	1
        total	575
        entry	2557010

        1
        id	13148025
        rank	2
        last_rank	5
        rank_sort	2
        total	572
        entry	2415205
*/
type LeagueParticipants struct {
	LeagueStandings LeagueStandings `json:"standings"`
}
type LeagueStandings struct {
	LeagueResults []LeagueResults `json:"results"`
}
type LeagueResults struct {
	Entry int64 `json:"entry"`
}

func makeRequest(fantasyMain *fantasyMain, URL string) []byte {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "pg-fpl")

	resp, err := fantasyMain.httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	return body
}

func getTeamInfoForParticipant(participantNumber int64, gameweek int, playerOccurance map[string]int, fantasyMain *fantasyMain) error {
	teamURL := fmt.Sprintf(teamURL, participantNumber, gameweek)

	response := makeRequest(fantasyMain, teamURL)
	ParticipantTeamInfo := new(ParticipantTeamInfo)
	err := json.Unmarshal(response, &ParticipantTeamInfo)
	if err != nil {
		return err
	}

	for _, player := range ParticipantTeamInfo.TeamPlayers {
		playerOccurance[fantasyMain.playerMap[player.Element]]++
	}
	return nil
}

func getPlayerMapping(fantasyMain *fantasyMain) int {
	response := makeRequest(fantasyMain, allPlayersURL)
	allPlayers := new(AllPlayers)
	err := json.Unmarshal(response, &allPlayers)
	if err != nil {
		panic(err.Error())
	}

	for _, player := range allPlayers.Players {
		fantasyMain.playerMap[player.ID] = player.WebName
	}

	fmt.Printf("Fetched data of %v premier league players \n", strconv.Itoa(len(fantasyMain.playerMap)))
	return len(fantasyMain.playerMap)

}

func getParticipantsInLeague(fantasyMain *fantasyMain, leagueCode int) int {
	participantsURL := fmt.Sprintf(participantsURL, leagueCode)

	response := makeRequest(fantasyMain, participantsURL)
	leagueParticipants := new(LeagueParticipants)
	err := json.Unmarshal(response, &leagueParticipants)
	if err != nil {
		panic(err.Error())
	}

	for _, participant := range leagueParticipants.LeagueStandings.LeagueResults {
		fantasyMain.leagueParticipants = append(fantasyMain.leagueParticipants, participant.Entry)
	}

	fmt.Printf("Fetched %v participants in league", strconv.Itoa(len(fantasyMain.leagueParticipants)))
	return len(fantasyMain.leagueParticipants)
}

func writeToFile(fantasyMain *fantasyMain, leagueCode int) {
	fmt.Println("Writing to file ...")

	fileName := fmt.Sprintf(csvFileName, time.Now().Format("2006-01-02"), leagueCode)
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	numOfGameweeks := len(fantasyMain.playerOccurances)
	//Headers
	var record []string
	record = append(record, "Player")
	for gameweekNum := 1; gameweekNum <= numOfGameweeks; gameweekNum++ {
		record = append(record, fmt.Sprintf("Gameweek %v", gameweekNum))
	}

	err = writer.Write(record)
	if err != nil {
		panic(err)
	}

	allPlayers := fantasyMain.playerOccurances[numOfGameweeks]

	for player := range allPlayers {

		var record []string
		record = append(record, string(player))

		for gameweekNum := 1; gameweekNum <= numOfGameweeks; gameweekNum++ {
			playerOccuranceForGameweek := fantasyMain.playerOccurances[gameweekNum]
			record = append(record, strconv.Itoa(playerOccuranceForGameweek[player]))
		}

		err := writer.Write(record)
		if err != nil {
			panic(err)
		}
	}
}

type greeterServer struct{}

// SayHello implements helloworld.GreeterServer
func (s *greeterServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

type fplServer struct{}

func createFantasyObject() *fantasyMain {
	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	fantasyMain := &fantasyMain{
		httpClient:       httpClient,
		playerMap:        make(map[int64]string),
		playerOccurances: make(map[int]map[string]int),
	}

	return fantasyMain
}

func (s *fplServer) GetNumberOfPlayers(context.Context, *pb.NumPlayerRequest) (*pb.NumPlayers, error) {
	numPlayersInFPL := getPlayerMapping(createFantasyObject())
	return &pb.NumPlayers{NumPlayers: int64(numPlayersInFPL)}, nil
}

func (s *fplServer) GetParticipantsInLeague(cxt context.Context, leagueCode *pb.LeagueCode) (*pb.NumParticipants, error) {
	numParticipants := getParticipantsInLeague(createFantasyObject(), int(leagueCode.LeagueCode))
	return &pb.NumParticipants{NumParticipants: int64(numParticipants)}, nil
}

func (s *fplServer) GetDataForGameweek(req *pb.GameweekReq, stream pb.FPL_GetDataForGameweekServer) error {
	fplObject := createFantasyObject()
	getPlayerMapping(fplObject)
	getParticipantsInLeague(fplObject, int(req.LeagueCode))

	playerOccuranceForGameweek := make(map[string]int)
	fmt.Printf("Fetching data for gameweek %v\n", req.Gameweek)

	for _, participant := range fplObject.leagueParticipants[0:10] {
		err := getTeamInfoForParticipant(participant, int(req.Gameweek), playerOccuranceForGameweek, fplObject)
		if err != nil {
			break
		}
	}
	if len(playerOccuranceForGameweek) > 0 {
		// playerOccuranceForGameweekMap := make(map[int]map[string]int)
		// playerOccuranceForGameweekMap[gameweek] = playerOccuranceForGameweek
		// playerOccuranceChan <- playerOccuranceForGameweekMap
		for player, occurance := range playerOccuranceForGameweek {
			if err := stream.Send(&pb.PlayerOccuranceData{
				PlayerName:                 player,
				PlayerOccuranceForGameweek: int32(occurance),
			},
			); err != nil {
				return err
			}
		}
	}
	// return &pb.NumParticipants{NumParticipants: int64(numParticipants)}, nil
	return nil
}

func startgRPCServer() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Creates a new gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterGreeterServer(grpcServer, &greeterServer{})
	pb.RegisterFPLServer(grpcServer, &fplServer{})
	fmt.Println("started grpc server ...")
	grpcServer.Serve(lis)
}

func main() {
	//	var wg sync.WaitGroup
	fmt.Println("Starting main program")
	startgRPCServer()

	// start := time.Now()
	// defer func() {
	// 	fmt.Printf("Took %v to fetch all data\n", time.Since(start))
	// }()

	// playerOccuranceChan := make(chan map[int]map[string]int)

	// gameweekMax := 38
	//leagueCode := 313

	// getPlayerMapping(fantasyMain)
	// getParticipantsInLeague(fantasyMain, leagueCode)

	// for gameweek := 1; gameweek <= gameweekMax; gameweek++ {
	// 	wg.Add(1)
	// 	go func(gameweek int) {
	// 		playerOccuranceForGameweek := make(map[string]int)
	// 		fmt.Printf("Fetching data for gameweek %v\n", gameweek)

	// 		for _, participant := range fantasyMain.leagueParticipants[0:10] {
	// 			err := getTeamInfoForParticipant(participant, gameweek, playerOccuranceForGameweek, fantasyMain)
	// 			if err != nil {
	// 				break
	// 			}
	// 		}
	// 		if len(playerOccuranceForGameweek) > 0 {
	// 			playerOccuranceForGameweekMap := make(map[int]map[string]int)
	// 			playerOccuranceForGameweekMap[gameweek] = playerOccuranceForGameweek
	// 			playerOccuranceChan <- playerOccuranceForGameweekMap
	// 		}
	// 		wg.Done()
	// 	}(gameweek)
	// }

	// go func() {
	// 	wg.Wait()
	// 	close(playerOccuranceChan)
	// }()

	// for playerOccuranceForGameweekMap := range playerOccuranceChan {
	// 	for gameweekNum, playerOccuranceForGameweek := range playerOccuranceForGameweekMap {
	// 		fmt.Printf("Data fetched for gameweek %v!\n", gameweekNum)
	// 		fantasyMain.playerOccurances[gameweekNum] = playerOccuranceForGameweek
	// 	}
	// }
	// writeToFile(fantasyMain, leagueCode)
}
