package steamquery

import (
	"errors"
	"time"

	"github.com/rumblefrog/go-a2s"
)

type DayZQuery struct {
	ServerName string `json:"name"`
	Map        string `json:"map"`
	Folder     string `json:"folder"`
	Game       string `json:"game"`
	MaxPlayers int    `json:"maxplayers"`
	Players    int    `json:"players"`
	ServerType int    `json:"type"`
	ServerOS   int    `json:"os"`
	VacEnabled bool   `json:"vac"`
	Version    string `json:"version"`
	Port       int    `json:"port"`
}

func QueryDayZ(ip string, queryport string, timeout time.Duration) (*DayZQuery, error) {
	client, err := a2s.NewClient(
		ip+":"+queryport,
		a2s.TimeoutOption(timeout),
	)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	info, err := client.QueryInfo()
	if err != nil {
		return nil, err
	}
	if info.ExtendedServerInfo.GameID != 221100 {
		return nil, errors.New("invalid game id. server must be 221100.")
	}
	result := &DayZQuery{
		ServerName: info.Name,
		Map:        info.Map,
		Folder:     info.Folder,
		Game:       info.Game,
		MaxPlayers: int(info.MaxPlayers),
		Players:    int(info.Players),
		ServerType: int(info.ServerType),
		ServerOS:   int(info.ServerOS),
		VacEnabled: info.VAC,
		Version:    info.Version,
		Port:       int(info.ExtendedServerInfo.Port),
	}
	return result, nil
}
