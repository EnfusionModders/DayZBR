package leaderboard

import (
	"encoding/json"
	"io"

	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
)

type SubmitMatchRequest struct {
	ServerID  string              `json:"server_id"`
	MatchData database.BRRawMatch `json:"match_data"`
}

type SteamInfoRequest struct {
	SteamID string `json:"steamid"`
}
type PlayerDataRequest struct {
	SteamID string `json:"steamid"`
}
type RankRequest struct {
	Rating float64 `json:"rating"`
}
type MatchDataRequest struct {
	MatchID            string `json:"matchid"`
	IsLeaderboardMatch bool   `json:"ldbmatch"`
}

func ParseSubmitMatchRequest(r io.ReadCloser) (*SubmitMatchRequest, error) {
	var request SubmitMatchRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseSteamInfoRequest(r io.ReadCloser) (*SteamInfoRequest, error) {
	var request SteamInfoRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParsePlayerDataRequest(r io.ReadCloser) (*PlayerDataRequest, error) {
	var request PlayerDataRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseMatchDataRequest(r io.ReadCloser) (*MatchDataRequest, error) {
	var request MatchDataRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseRankRequest(r io.ReadCloser) (*RankRequest, error) {
	var request RankRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
