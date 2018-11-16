package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
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

func GetTeamInfoForParticipant(participantNumber int64, gameweek int, playerOccurance map[string]int, myFPLServer FPLServer) error {
	teamURL := fmt.Sprintf(teamURL, participantNumber, gameweek)

	response, err := myFPLServer.MakeRequest(teamURL)
	if err != nil {
		return err
	}

	ParticipantTeamInfo := new(ParticipantTeamInfo)
	err = json.Unmarshal(response, &ParticipantTeamInfo)
	if err != nil {
		fmt.Println("error ")
		return errors.Errorf("error unmarshalling response URL %v for GetTeamInfoForParticipant: %v", teamURL, err)
	}

	playerMap := myFPLServer.GetPlayerMap()
	for _, player := range ParticipantTeamInfo.TeamPlayers {
		playerOccurance[playerMap[player.Element]]++
	}
	return nil
}

func GetPlayerMapping(myFPLServer *MyFPLServer) (int, error) {

	response, err := makeRequest(myFPLServer, allPlayersURL)
	if err != nil {
		return 0, err
	}

	allPlayers := new(AllPlayers)
	err = json.Unmarshal(response, &allPlayers)
	if err != nil {
		return 0, errors.Errorf("error unmarshalling response for GetPlayerMapping : %v", err)
	}

	for _, player := range allPlayers.Players {
		myFPLServer.playerMap[player.ID] = player.WebName
	}

	fmt.Printf("Fetched data of %v premier league players \n", strconv.Itoa(len(myFPLServer.playerMap)))
	return len(myFPLServer.playerMap), nil

}

func GetParticipantsInLeague(myFPLServer *MyFPLServer, leagueCode int) (int, error) {
	participantsURL := fmt.Sprintf(participantsURL, leagueCode)

	response, err := makeRequest(myFPLServer, participantsURL)
	if err != nil {
		return 0, err
	}

	leagueParticipants := new(LeagueParticipants)
	err = json.Unmarshal(response, &leagueParticipants)
	if err != nil {
		return 0, errors.Errorf("could not parse response for GetParticipantsInLeague: %v", err)
	}

	for _, participant := range leagueParticipants.LeagueStandings.LeagueResults {
		myFPLServer.leagueParticipants = append(myFPLServer.leagueParticipants, participant.Entry)
	}

	fmt.Printf("Fetched %v participants in league", strconv.Itoa(len(myFPLServer.leagueParticipants)))
	return len(myFPLServer.leagueParticipants), nil
}

func WriteToFile(myFPLServer *MyFPLServer, leagueCode int) (string, error) {
	fmt.Println("Writing to file ...")

	fileName := fmt.Sprintf(csvFileName, time.Now().Format("2006-01-02"), leagueCode)
	file, err := os.Create(fileName)
	if err != nil {
		return "", errors.Errorf("error creating file %v : %v", fileName, err)
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
		return "", errors.Errorf("error writing to file %v: %v", fileName, err)
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
			return "", errors.Errorf("error writing to file %v : %v", fileName, err)
		}
	}
	return fileName, nil
}

func makeRequest(myFPLServer *MyFPLServer, URL string) ([]byte, error) {

	var err error
	customErr := errors.Errorf("error with request to %v : %v", URL, err)

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, customErr
	}

	req.Header.Set("User-Agent", "pg-fpl")

	resp, err := myFPLServer.httpClient.Do(req)
	if err != nil {
		return nil, customErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, customErr
	}

	return body, nil
}
