package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	teamURL         = "https://fantasy.premierleague.com/drf/entry/%v/event/%v/picks"
	allPlayersURL   = "https://fantasy.premierleague.com/drf/bootstrap-static"
	participantsURL = "https://fantasy.premierleague.com/drf/leagues-classic-standings/%v?phase=1&le-page=1&ls-page=1"
	csvFileName     = "temp-%v-%v.csv"
	gameweekMax     = 38
)

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

func makeRequest(myFPLServer *MyFPLServer, URL string) []byte {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "pg-fpl")

	resp, err := myFPLServer.httpClient.Do(req)
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

func GetTeamInfoForParticipant(participantNumber int64, gameweek int, playerOccurance map[string]int, myFPLServer *MyFPLServer) error {
	teamURL := fmt.Sprintf(teamURL, participantNumber, gameweek)

	response := makeRequest(myFPLServer, teamURL)
	ParticipantTeamInfo := new(ParticipantTeamInfo)
	err := json.Unmarshal(response, &ParticipantTeamInfo)
	if err != nil {
		return err
	}

	for _, player := range ParticipantTeamInfo.TeamPlayers {
		playerOccurance[myFPLServer.playerMap[player.Element]]++
	}
	return nil
}

func GetPlayerMapping(myFPLServer *MyFPLServer) int {
	response := makeRequest(myFPLServer, allPlayersURL)
	allPlayers := new(AllPlayers)
	err := json.Unmarshal(response, &allPlayers)
	if err != nil {
		panic(err.Error())
	}

	for _, player := range allPlayers.Players {
		myFPLServer.playerMap[player.ID] = player.WebName
	}

	fmt.Printf("Fetched data of %v premier league players \n", strconv.Itoa(len(myFPLServer.playerMap)))
	return len(myFPLServer.playerMap)

}

func GetParticipantsInLeague(myFPLServer *MyFPLServer, leagueCode int) int {
	participantsURL := fmt.Sprintf(participantsURL, leagueCode)

	response := makeRequest(myFPLServer, participantsURL)
	leagueParticipants := new(LeagueParticipants)
	err := json.Unmarshal(response, &leagueParticipants)
	if err != nil {
		panic(err.Error())
	}

	for _, participant := range leagueParticipants.LeagueStandings.LeagueResults {
		myFPLServer.leagueParticipants = append(myFPLServer.leagueParticipants, participant.Entry)
	}

	fmt.Printf("Fetched %v participants in league", strconv.Itoa(len(myFPLServer.leagueParticipants)))
	return len(myFPLServer.leagueParticipants)
}

func WriteToFile(myFPLServer *MyFPLServer, leagueCode int) string {
	fmt.Println("Writing to file ...")

	fileName := fmt.Sprintf(csvFileName, time.Now().Format("2006-01-02"), leagueCode)
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	numOfGameweeks := len(myFPLServer.playerOccurances)
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

	allPlayers := myFPLServer.playerOccurances[numOfGameweeks]

	for player := range allPlayers {

		var record []string
		record = append(record, string(player))

		for gameweekNum := 1; gameweekNum <= numOfGameweeks; gameweekNum++ {
			playerOccuranceForGameweek := myFPLServer.playerOccurances[gameweekNum]
			record = append(record, strconv.Itoa(playerOccuranceForGameweek[player]))
		}

		err := writer.Write(record)
		if err != nil {
			panic(err)
		}
	}
	return fileName
}
