package server_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	grpc_fpl "github.com/go-fantasy/fpl/grpc"
	"github.com/go-fantasy/fpl/mock"
	"github.com/go-fantasy/fpl/server"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type TestServer struct {
	suite.Suite
	playerMap   map[int64]string
	myServer    *server.MyFPLServer
	mockScraper *mock_server.MockScraper
	ctx         context.Context
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(TestServer))
}

//Run once before all tests
func (suite *TestServer) SetupSuite() {

	playerMap := make(map[int64]string)
	playerMap[267] = "Messi"
	playerMap[247] = "Ronaldo"
	playerMap[454] = "Salah"

	mockCtrl := gomock.NewController(suite.T())
	defer mockCtrl.Finish()

	testObj := mock_server.NewMockScraper(mockCtrl)
	myFPLServer := &server.MyFPLServer{
		PlayerMap:        make(map[int64]string),
		PlayerOccurances: make(map[int]map[string]int),
		Scraper:          testObj,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	suite.playerMap = playerMap
	suite.myServer = myFPLServer
	suite.mockScraper = testObj
	suite.ctx = ctx
}

//Run once before each test
func (suite *TestServer) SetupTest() {

}

func (s *TestServer) TestGetNumberOfPlayers() {
	t := s.T()

	s.mockScraper.EXPECT().GetPlayerMapping().Return(s.playerMap, nil).Times(1)

	numPlayers, err := s.myServer.GetNumberOfPlayers(s.ctx, &grpc_fpl.NumPlayerRequest{})
	assert.Nil(t, err)

	//log.Printf("There are %v players in fpl!", numPlayers.NumPlayers)

	assert.Equal(t, len(s.playerMap), int(numPlayers.NumPlayers))
	// for _, value := range playerMap {
	// 	assert.Equal(t, playerOccuranceForGameweek[value], 1, "Values not matching for %v", value)
	// }
}

func (s *TestServer) TestGetDataForGameweek() {
	t := s.T()

	playerOccuranceForGameweek := make(map[string]int)
	expectedOccurance := 2

	for _, player := range s.playerMap {
		playerOccuranceForGameweek[player] = expectedOccurance
	}
	s.mockScraper.EXPECT().GetPlayerMapping().Return(s.playerMap, nil).Times(1)
	s.mockScraper.EXPECT().GetParticipantsInLeague(gomock.Any()).Return(&[]int64{1, 2}, nil).Times(1)
	s.mockScraper.EXPECT().GetTeamInfoForParticipant(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(playerOccuranceForGameweek, nil).Times(1)
	playerOccurance, err := s.myServer.GetDataForGameweek(s.ctx, &grpc_fpl.GameweekReq{LeagueCode: 1, Gameweek: 1})
	assert.Nil(t, err, "Error %v was supposed to be nil ", err)

	for player, occurance := range playerOccurance.PlayerOccurance {
		log.Printf("Player %v was selected by %v players!", player, occurance)
		assert.Equal(t, int(occurance), expectedOccurance, "Player %v was supposed to be selected 2 times!", player)
	}

}

type mockStream struct {
	grpc.ServerStream
}

func (x *mockStream) Send(m *grpc_fpl.AllGameweekData) error {
	log.Println("Calling mock send function!!")
	return nil
}

func (s *TestServer) TestGetDataForAllGameweeks() {
	t := s.T()

	leagueCode := int64(1)
	s.mockScraper.EXPECT().GetPlayerMapping().Return(s.playerMap, nil).Times(1)
	s.mockScraper.EXPECT().GetParticipantsInLeague(gomock.Any()).Return(&[]int64{1, 2}, nil).Times(1)

	playerOccuranceForGameweek := make(map[string]int)
	for _, player := range s.playerMap {
		playerOccuranceForGameweek[player] = 2
	}
	s.mockScraper.EXPECT().GetTeamInfoForParticipant(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(playerOccuranceForGameweek, nil).AnyTimes()

	//Create temp file
	fileName, err := getTempFile()
	assert.NotNil(t, fileName)
	assert.Nil(t, err)

	s.mockScraper.EXPECT().WriteToFile(gomock.Any(), gomock.Any()).
		Do(func(playerOccurances map[int]map[string]int, leagueCode int) {
			assert.Equal(t, leagueCode, leagueCode, "League code not matching!!")
			assert.Equal(t, len(playerOccurances), server.GameweekMax, "Length of playerOccurances not matching! %v", len(playerOccurances))

			for gameweek := 1; gameweek <= server.GameweekMax; gameweek++ {
				assert.Equal(t, playerOccurances[gameweek], playerOccuranceForGameweek, "playerOccuranceForGameweek not matching! %v ", gameweek)
			}

		}).Return(fileName, nil).Times(1)

	err = s.myServer.GetDataForAllGameweeks(&grpc_fpl.LeagueCode{LeagueCode: leagueCode}, &mockStream{})
	assert.Nil(t, err, "Error %v was supposed to be nil ", err)
	_, err = os.Open(fileName)
	assert.NotNil(t, err, "File %v should have been deleted! ", fileName)

}

func getTempFile() (string, error) {
	scraper := &server.MyFPLScraper{
		Client: &server.MyFPLClient{
			HttpClient: nil,
		},
	}
	playerOccuranceForAllGameweeks := make(map[int]map[string]int)
	fileName, err := scraper.WriteToFile(playerOccuranceForAllGameweeks, 1)
	return fileName, err
}
