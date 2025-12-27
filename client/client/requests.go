package client

import (
	"encoding/json"
	"io"
)

type StartRequest struct {
	SteamId string `json:"steamid"`
	Name    string `json:"name"`
}
type MatchMakeRequest struct {
	PlayerId string `json:"id"`
	Region   string `json:"region"`
	Version  string `json:"version"`
}
type GetPlayerRequest struct {
	PlayerId string `json:"id"`
}
type GetMatchRequest struct {
	MatchId string `json:"id"`
}
type GetServerRequest struct {
	ServerId string `json:"id"`
}

func ParseStartRequest(r io.ReadCloser) (*StartRequest, error) {
	var request StartRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseMatchMakeRequest(r io.ReadCloser) (*MatchMakeRequest, error) {
	var request MatchMakeRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseGetPlayerRequest(r io.ReadCloser) (*GetPlayerRequest, error) {
	var request GetPlayerRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseGetMatchRequest(r io.ReadCloser) (*GetMatchRequest, error) {
	var request GetMatchRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseGetServerRequest(r io.ReadCloser) (*GetServerRequest, error) {
	var request GetServerRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
