package server

import (
	"net/http"

	"github.com/go-fantasy/fpl/grpc"
)

//FPLServer is the main interface for the application
type FPLServer interface {
	grpc.FPLServer
	Start(string) error
}

//Scraper is the main scraping interface for the FPL app
type Scraper interface {
	GetTeamInfoForParticipant(map[int64]string, int, *[]int64) (map[string]int, error)
	GetPlayerMapping() (map[int64]string, error)
	GetParticipantsInLeague(int) (*[]int64, error)
	WriteToFile(map[int]map[string]int, int) (string, error)
}

//Client is the interface for making API calls to FPL site
type Client interface {
	MakeRequest(string) ([]byte, error)
}

//MyFPLServer is my implementation of the FPL server
type MyFPLServer struct {
	PlayerMap          map[int64]string
	LeagueParticipants *[]int64
	PlayerOccurances   map[int]map[string]int
	Scraper            Scraper
}

//MyFPLScraper is my implementation of the FPL server scraper interface
type MyFPLScraper struct {
	Client
}

//MyFPLClient is my implementation of the FPL client interface
type MyFPLClient struct {
	HttpClient *http.Client
}
