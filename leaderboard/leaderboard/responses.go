package leaderboard

import "gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"

type GenericResponse struct {
	Success int         `json:"success"`
	Error   string      `json:"error"`
	Data    interface{} `json:"data"`
}

type SteamInfoResponse struct {
	ProfileName string `json:"profilename"`
	ProfilePic  string `json:"profilepic"`
	ProfileUrl  string `json:"profileurl"`
}
type GlobalDataResponse struct {
	Players []database.LeaderboardPlayer   `json:"top"`
	Data    database.GlobalLeaderboardData `json:"data"`
}
type RankResponse struct {
	Rank int `json:"rank"`
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
