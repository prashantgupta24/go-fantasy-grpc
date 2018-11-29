package server_test

import (
	"fmt"
	"testing"

	"github.com/go-fantasy/fpl/mock"
	"github.com/go-fantasy/fpl/server"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetTeamInfoForParticipant(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	testObj := mock_server.NewMockClient(mockCtrl)
	playerMap := make(map[int64]string)
	playerMap[267] = "Messi"
	playerMap[247] = "Ronaldo"
	playerMap[454] = "Salah"

	b1 := `{  
   "active_chip":"",
   "automatic_subs":[  

   ],
   "entry_history":{  
      "id":1,
      "movement":"new",
      "points":99,
      "total_points":99,
      "rank":16627,
      "rank_sort":16627,
      "overall_rank":16627,
      "targets":null,
      "event_transfers":0,
      "event_transfers_cost":0,
      "value":1000,
      "points_on_bench":14,
      "bank":0,
      "entry":1,
      "event":1
   },
   "event":{  
      "id":1,
      "name":"Gameweek 1",
      "deadline_time":"2018-08-10T18:00:00Z",
      "average_entry_score":53,
      "finished":true,
      "data_checked":true,
      "highest_scoring_entry":890626,
      "deadline_time_epoch":1533924000,
      "deadline_time_game_offset":3600,
      "deadline_time_formatted":"10 Aug 19:00",
      "highest_score":137,
      "is_previous":false,
      "is_current":false,
      "is_next":false
   },
   "picks":[  
      {  
         "element":454,
         "position":1,
         "is_captain":false,
         "is_vice_captain":false,
         "multiplier":1
      },
      {  
         "element":267,
         "position":2,
         "is_captain":false,
         "is_vice_captain":false,
         "multiplier":1
      },
      {  
         "element":247,
         "position":3,
         "is_captain":false,
         "is_vice_captain":false,
         "multiplier":1
      }
   ]
}`

	b2 := `{  
   "active_chip":"",
   "automatic_subs":[  

   ],
   "entry_history":{  
      "id":1,
      "movement":"new",
      "points":99,
      "total_points":99,
      "rank":16627,
      "rank_sort":16627,
      "overall_rank":16627,
      "targets":null,
      "event_transfers":0,
      "event_transfers_cost":0,
      "value":1000,
      "points_on_bench":14,
      "bank":0,
      "entry":1,
      "event":1
   },
   "event":{  
      "id":1,
      "name":"Gameweek 1",
      "deadline_time":"2018-08-10T18:00:00Z",
      "average_entry_score":53,
      "finished":true,
      "data_checked":true,
      "highest_scoring_entry":890626,
      "deadline_time_epoch":1533924000,
      "deadline_time_game_offset":3600,
      "deadline_time_formatted":"10 Aug 19:00",
      "highest_score":137,
      "is_previous":false,
      "is_current":false,
      "is_next":false
   },
   "picks":[  
      {  
         "element":267,
         "position":2,
         "is_captain":false,
         "is_vice_captain":false,
         "multiplier":1
      },
      {  
         "element":247,
         "position":3,
         "is_captain":false,
         "is_vice_captain":false,
         "multiplier":1
      }
   ]
}`
	firstcall := testObj.EXPECT().MakeRequest(gomock.Any()).Do(func(s string) {
		fmt.Printf("Calling MakeRequest with %v url \n\n", s)
	}).Return([]byte(b1), nil).Times(1)

	testObj.EXPECT().MakeRequest(gomock.Any()).Do(func(s string) {
		fmt.Printf("Calling MakeRequest with %v url \n\n", s)
	}).Return([]byte(b2), nil).After(firstcall).Times(1)

	testScraper := &server.MyFPLScraper{
		Client: testObj,
	}
	//participantsInLeague := []int64{2575352, 3614956, 8995, 8450}
	playerOccuranceForGameweek, err := testScraper.GetTeamInfoForParticipant(playerMap, 1, &[]int64{1, 2})
	assert.Nil(t, err)

	for _, player := range playerMap {
		if player == "Salah" {
			assert.Equal(t, playerOccuranceForGameweek[player], 1, "Values not matching for %v", player)
		} else {
			assert.Equal(t, playerOccuranceForGameweek[player], 2, "Values not matching for %v", player)
		}
	}
}

func TestGetPlayerMapping(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	testObj := mock_server.NewMockClient(mockCtrl)

	b := `{
  "phases": [
  ],
  "elements": [
    {
      "id": 1,
      "photo": "11334.jpg",
      "web_name": "Cech",
      "team_code": 3,
      "status": "a",
      "code": 11334,
      "first_name": "Petr",
      "second_name": "Cech",
      "squad_number": 1,
      "news": "",
      "now_cost": 50,
      "news_added": "2018-09-29T17:31:14Z",
      "chance_of_playing_this_round": 100,
      "chance_of_playing_next_round": 100,
      "value_form": "0.0",
      "value_season": "4.8",
      "cost_change_start": 0,
      "cost_change_event": 0,
      "cost_change_start_fall": 0,
      "cost_change_event_fall": 0,
      "in_dreamteam": false,
      "dreamteam_count": 0,
      "selected_by_percent": "1.4",
      "form": "0.0",
      "transfers_out": 110271,
      "transfers_in": 78068,
      "transfers_out_event": 1543,
      "transfers_in_event": 259,
      "loans_in": 0,
      "loans_out": 0,
      "loaned_in": 0,
      "loaned_out": 0,
      "total_points": 24,
      "event_points": 0,
      "points_per_game": "3.4",
      "ep_this": "0.5",
      "ep_next": "0.5",
      "special": false,
      "minutes": 585,
      "goals_scored": 0,
      "assists": 0,
      "clean_sheets": 1,
      "goals_conceded": 9,
      "own_goals": 0,
      "penalties_saved": 0,
      "penalties_missed": 0,
      "yellow_cards": 0,
      "red_cards": 0,
      "saves": 27,
      "bonus": 3,
      "bps": 130,
      "influence": "205.0",
      "creativity": "0.0",
      "threat": "0.0",
      "ict_index": "20.4",
      "ea_index": 0,
      "element_type": 1,
      "team": 1
    },
    {
      "id": 2,
      "photo": "80201.jpg",
      "web_name": "Leno",
      "team_code": 3,
      "status": "a",
      "code": 80201,
      "first_name": "Bernd",
      "second_name": "Leno",
      "squad_number": 19,
      "news": "",
      "now_cost": 48,
      "news_added": null,
      "chance_of_playing_this_round": null,
      "chance_of_playing_next_round": null,
      "value_form": "0.5",
      "value_season": "3.1",
      "cost_change_start": -2,
      "cost_change_event": 0,
      "cost_change_start_fall": 2,
      "cost_change_event_fall": 0,
      "in_dreamteam": false,
      "dreamteam_count": 0,
      "selected_by_percent": "1.7",
      "form": "2.5",
      "transfers_out": 79984,
      "transfers_in": 41939,
      "transfers_out_event": 1227,
      "transfers_in_event": 1684,
      "loans_in": 0,
      "loans_out": 0,
      "loaned_in": 0,
      "loaned_out": 0,
      "total_points": 15,
      "event_points": 4,
      "points_per_game": "2.5",
      "ep_this": "3.0",
      "ep_next": "3.0",
      "special": false,
      "minutes": 495,
      "goals_scored": 0,
      "assists": 0,
      "clean_sheets": 0,
      "goals_conceded": 6,
      "own_goals": 0,
      "penalties_saved": 0,
      "penalties_missed": 0,
      "yellow_cards": 0,
      "red_cards": 0,
      "saves": 16,
      "bonus": 1,
      "bps": 85,
      "influence": "128.0",
      "creativity": "0.0",
      "threat": "0.0",
      "ict_index": "12.8",
      "ea_index": 0,
      "element_type": 1,
      "team": 1
    }
]
}`

	testObj.EXPECT().MakeRequest(gomock.Any()).Return([]byte(b), nil).Times(1)
	testScraper := &server.MyFPLScraper{
		Client: testObj,
	}
	playerMap, err := testScraper.GetPlayerMapping()
	assert.Equal(t, len(playerMap), 2)
	assert.Nil(t, err)

}

func TestGetParticipantsInLeague(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	testObj := mock_server.NewMockClient(mockCtrl)

	b := `{  
   "standings":{  
      "has_next":true,
      "number":1,
      "results":[  
         {  
            "id":14105046,
            "entry_name":"A's team",
            "event_total":65,
            "player_name":"A",
            "movement":"same",
            "own_entry":false,
            "rank":1,
            "last_rank":1,
            "rank_sort":1,
            "total":918,
            "entry":2575352,
            "league":313,
            "start_event":1,
            "stop_event":38
         },
         {  
            "id":20781604,
            "entry_name":"B's team",
            "event_total":68,
            "player_name":"B",
            "movement":"same",
            "own_entry":false,
            "rank":2,
            "last_rank":2,
            "rank_sort":2,
            "total":908,
            "entry":3614956,
            "league":313,
            "start_event":1,
            "stop_event":38
         },
         {  
            "id":229597,
            "entry_name":"C's team",
            "event_total":69,
            "player_name":"C",
            "movement":"up",
            "own_entry":false,
            "rank":3,
            "last_rank":7,
            "rank_sort":3,
            "total":899,
            "entry":48995,
            "league":313,
            "start_event":1,
            "stop_event":38
         },
         {  
            "id":40188,
            "entry_name":"D's team",
            "event_total":67,
            "player_name":"D",
            "movement":"same",
            "own_entry":false,
            "rank":4,
            "last_rank":4,
            "rank_sort":4,
            "total":898,
            "entry":8450,
            "league":313,
            "start_event":1,
            "stop_event":38
         }
      ]
   }
}`
	testObj.EXPECT().MakeRequest(gomock.Any()).Return([]byte(b), nil).Times(1)

	testScraper := &server.MyFPLScraper{
		Client: testObj,
	}
	leagueParticipants, err := testScraper.GetParticipantsInLeague(1)
	//fmt.Println(*leagueParticipants)

	assert.Nil(t, err)
	assert.Equal(t, 4, len(*leagueParticipants))
}
