package server

import (
	"net/http"
)

//FPLServer is the main interface for the application
type FPLServer interface {
	Start(string) error
}

//MyFPLServer holds the building block for the application
type MyFPLServer struct {
	httpClient         *http.Client
	playerMap          map[int64]string
	leagueParticipants []int64
	playerOccurances   map[int]map[string]int
}
