package client

import (
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/steam-query/steamquery"
)

type MatchMakeQuery struct {
	Query  *steamquery.DayZQuery
	Server *database.BRServer
}
