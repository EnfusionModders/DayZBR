package server

import (
	"encoding/json"
	"io"
)

type OnStartRequest struct {
	QueryPort     string `json:"query_port"`
	ServerVersion string `json:"server_version"`
	ServerIP      string `json:"server_ip"`
}
type OnFinishRequest struct {
	Winner    string `json:"winner"`
	QueryPort string `json:"query_port"`
	ServerIP  string `json:"server_ip"`
}
type SetLockRequest struct {
	Lock      uint8  `json:"lock"`
	QueryPort string `json:"query_port"`
	ServerIP  string `json:"server_ip"`
}

func ParseOnStartRequest(r io.ReadCloser) (*OnStartRequest, error) {
	var request OnStartRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseOnFinishRequest(r io.ReadCloser) (*OnFinishRequest, error) {
	var request OnFinishRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
func ParseSetLockRequest(r io.ReadCloser) (*SetLockRequest, error) {
	var request SetLockRequest
	err := json.NewDecoder(r).Decode(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
