package client

import "gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"

type GenericResponse struct {
	Success int         `json:"success"`
	Error   string      `json:"error"`
	Data    interface{} `json:"data"`
}
type MatchmakeResponse struct {
	Wait   uint8             `json:"wait"`
	Server database.BRServer `json:"server"`
}
type StartResponse struct {
	Player database.BRPlayer `json:"player"`
	Region struct {
		Regions []string `json:"regions"`
	} `json:"region"`
}

func buildSuccessResponse(data interface{}) *GenericResponse {
	return &GenericResponse{
		Success: 1,
		Error:   "",
		Data:    data,
	}
}

func buildErrorResponse(message string) *GenericResponse {
	return &GenericResponse{
		Success: 0,
		Error:   message,
		Data:    nil,
	}
}
