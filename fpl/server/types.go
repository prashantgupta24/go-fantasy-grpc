package server

import (
	"net/http"

	"github.com/go-fantasy/fpl/grpc"
)

//FPLServer is the main interface for the application
type FPLServer interface {
	grpc.FPLServer
	Start(string) error
	MakeRequest(string) ([]byte, error)
	GetPlayerOccurances() map[int]map[string]int
	GetPlayerMap() map[int64]string
}

//MyFPLServer holds the building block for the application
type MyFPLServer struct {
	httpClient         *http.Client
	playerMap          map[int64]string
	leagueParticipants []int64
	playerOccurances   map[int]map[string]int
}
