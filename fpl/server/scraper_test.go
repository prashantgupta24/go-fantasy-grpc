package server

import (
	"fmt"
	"testing"

	"github.com/go-fantasy/fpl/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetTeamInfoForParticipant(t *testing.T) {
	// type args struct {
	// 	participantNumber int64
	// 	gameweek          int
	// 	playerOccurance   map[string]int
	// 	myFPLServer       *MyFPLServer
	// }
	// tests := []struct {
	// 	name    string
	// 	args    args
	// 	wantErr bool
	// }{
	// 	// TODO: Add test cases.
	// }
	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		if err := GetTeamInfoForParticipant(tt.args.participantNumber, tt.args.gameweek, tt.args.playerOccurance, tt.args.myFPLServer); (err != nil) != tt.wantErr {
	// 			t.Errorf("GetTeamInfoForParticipant() error = %v, wantErr %v", err, tt.wantErr)
	// 		}
	// 	})
	// }
	// _ = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
	// 	// Test request parameters
	// 	//equals(t, req.URL.String(), "/some/path")
	// 	fmt.Printf("values are : %v", req.URL.String())
	// 	// Send response to be tested
	// 	rw.Write([]byte(`OK`))
	// }))
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	testObj := mock_server.NewMockFPLServer(mockCtrl)
	playerMap := make(map[int64]string)
	playerMap[267] = "Messi"
	playerMap[247] = "Ronaldo"
	playerMap[454] = "Salah"

	b := `{"active_chip":"","automatic_subs":[],"entry_history":{"id":1,"movement":"new","points":99,"total_points":99,"rank":16627,"rank_sort":16627,"overall_rank":16627,"targets":null,"event_transfers":0,"event_transfers_cost":0,"value":1000,"points_on_bench":14,"bank":0,"entry":1,"event":1},"event":{"id":1,"name":"Gameweek 1","deadline_time":"2018-08-10T18:00:00Z","average_entry_score":53,"finished":true,"data_checked":true,"highest_scoring_entry":890626,"deadline_time_epoch":1533924000,"deadline_time_game_offset":3600,"deadline_time_formatted":"10 Aug 19:00","highest_score":137,"is_previous":false,"is_current":false,"is_next":false},"picks":[{"element":454,"position":1,"is_captain":false,"is_vice_captain":false,"multiplier":1},{"element":267,"position":2,"is_captain":false,"is_vice_captain":false,"multiplier":1},{"element":247,"position":3,"is_captain":false,"is_vice_captain":false,"multiplier":1}]}`
	teamURL := fmt.Sprintf(teamURL, 1, 1)
	testObj.EXPECT().MakeRequest(teamURL).Return([]byte(b), nil).Times(1)
	testObj.EXPECT().GetPlayerMap().Return(playerMap).Times(1)
	// myFPLServer := &MyFPLServer{
	// 	httpClient:       &http.Client{},
	// 	playerMap:        make(map[int64]string),
	// 	playerOccurances: make(map[int]map[string]int),
	// }
	playerOccuranceForGameweek := make(map[string]int)

	err := GetTeamInfoForParticipant(1, 1, playerOccuranceForGameweek, testObj)
	assert.Nil(t, err)

	for key, value := range playerOccuranceForGameweek {
		fmt.Printf("key %v and value %v\n", key, value)
	}
}
