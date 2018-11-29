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

/*
GameweekMax is the max gameweek to fetch upto
*/
const (
	teamURL         = "https://fantasy.premierleague.com/drf/entry/%v/event/%v/picks"
	allPlayersURL   = "https://fantasy.premierleague.com/drf/bootstrap-static"
	participantsURL = "https://fantasy.premierleague.com/drf/leagues-classic-standings/%v?phase=1&le-page=1&ls-page=1"
	csvFileName     = "temp-%v-%v.csv"
	GameweekMax     = 38
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

//GetTeamInfoForParticipant gets a map of players and their picks for a gameweek for all participants provided
func (s *MyFPLScraper) GetTeamInfoForParticipant(playerMap map[int64]string, gameweek int, topLeagueParticipants *[]int64) (map[string]int, error) {

	playerOccuranceForGameweek := make(map[string]int)
	for _, participant := range *topLeagueParticipants {
		teamURL := fmt.Sprintf(teamURL, participant, gameweek)

		response, err := s.MakeRequest(teamURL)
		if err != nil {
			return nil, err
		}

		ParticipantTeamInfo := new(ParticipantTeamInfo)
		err = json.Unmarshal(response, &ParticipantTeamInfo)
		if err != nil {
			//return nil, errors.Errorf("error unmarshalling response URL %v for GetTeamInfoForParticipant: %v", teamURL, err)
			break
		}

		for _, player := range ParticipantTeamInfo.TeamPlayers {
			playerOccuranceForGameweek[playerMap[player.Element]]++
		}
	}

	return playerOccuranceForGameweek, nil
}

func (s *MyFPLScraper) GetPlayerMapping() (map[int64]string, error) {

	response, err := s.MakeRequest(allPlayersURL)
	if err != nil {
		return nil, err
	}

	allPlayers := new(AllPlayers)
	err = json.Unmarshal(response, &allPlayers)
	if err != nil {
		return nil, errors.Errorf("error unmarshalling response for GetPlayerMapping : %v", err)
	}

	playerMap := make(map[int64]string)
	for _, player := range allPlayers.Players {
		playerMap[player.ID] = player.WebName
	}

	numPlayers := len(playerMap)
	fmt.Printf("Fetched data of %v premier league players \n", strconv.Itoa(numPlayers))
	return playerMap, nil

}

func (s *MyFPLScraper) GetParticipantsInLeague(leagueCode int) (*[]int64, error) {
	participantsURL := fmt.Sprintf(participantsURL, leagueCode)

	response, err := s.MakeRequest(participantsURL)
	if err != nil {
		return nil, err
	}

	leagueParticipants := new(LeagueParticipants)
	err = json.Unmarshal(response, &leagueParticipants)
	if err != nil {
		return nil, errors.Errorf("could not parse response for GetParticipantsInLeague for league %v: %v", leagueCode, err)
	}

	var leagueParticipantsData []int64
	for _, participant := range leagueParticipants.LeagueStandings.LeagueResults {
		leagueParticipantsData = append(leagueParticipantsData, participant.Entry)
	}

	fmt.Printf("Fetched %v participants in league\n", strconv.Itoa(len(leagueParticipantsData)))
	return &leagueParticipantsData, nil
}

func (s *MyFPLScraper) WriteToFile(playerOccurances map[int]map[string]int, leagueCode int) (string, error) {
	fmt.Println("Writing to file ...")

	fileName := fmt.Sprintf(csvFileName, time.Now().Format("2006-01-02"), leagueCode)
	file, err := os.Create(fileName)
	if err != nil {
		return "", errors.Errorf("error creating file %v : %v", fileName, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	numOfGameweeks := len(playerOccurances)
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

	allPlayersInLatestGameweek := playerOccurances[numOfGameweeks]

	for player := range allPlayersInLatestGameweek {

		var record []string
		record = append(record, string(player))

		for gameweekNum := 1; gameweekNum <= numOfGameweeks; gameweekNum++ {
			playerOccuranceForGameweek := playerOccurances[gameweekNum]
			record = append(record, strconv.Itoa(playerOccuranceForGameweek[player]))
		}

		err := writer.Write(record)
		if err != nil {
			return "", errors.Errorf("error writing to file %v : %v", fileName, err)
		}
	}
	return fileName, nil
}

func (client *MyFPLClient) MakeRequest(URL string) ([]byte, error) {

	var err error
	customErr := errors.Errorf("error with request to %v : %v", URL, err)

	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, customErr
	}

	req.Header.Set("User-Agent", "pg-fpl")

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return nil, customErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, customErr
	}

	return body, nil
}
