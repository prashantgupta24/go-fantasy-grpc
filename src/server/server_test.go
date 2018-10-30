package main

import (
	"testing"
)

func Test_writeToFile(t *testing.T) {
	type args struct {
		fantasyMain *fantasyMain
		leagueCode  int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeToFile(tt.args.fantasyMain, tt.args.leagueCode)
		})
	}
}

func Test_getParticipantsInLeague(t *testing.T) {
	type args struct {
		fantasyMain *fantasyMain
		leagueCode  int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getParticipantsInLeague(tt.args.fantasyMain, tt.args.leagueCode)
		})
	}
}

func Test_getPlayerMapping(t *testing.T) {
	type args struct {
		fantasyMain *fantasyMain
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getPlayerMapping(tt.args.fantasyMain)
		})
	}
}

func Test_getTeamInfoForParticipant(t *testing.T) {
	type args struct {
		participantNumber int64
		gameweek          int
		playerOccurance   map[string]int
		fantasyMain       *fantasyMain
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := getTeamInfoForParticipant(tt.args.participantNumber, tt.args.gameweek, tt.args.playerOccurance, tt.args.fantasyMain); (err != nil) != tt.wantErr {
				t.Errorf("getTeamInfoForParticipant() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
