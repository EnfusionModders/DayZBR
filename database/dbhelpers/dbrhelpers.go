package dbhelpers

import (
	"errors"

	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
)

func GetAllRegions(db *database.BattleRoyaleDB) ([]string, error) {
	if db == nil {
		return nil, errors.New("invalid database")
	}
	if !db.Connected {
		return nil, errors.New("database disconnected")
	}
	svrs, err := db.GetAllServers()
	if err != nil {
		return nil, err
	}

	included := make(map[string]bool, 0)
	var regions []string
	for _, svr := range svrs {
		if !included[svr.Region] {
			regions = append(regions, svr.Region)
			included[svr.Region] = true
		}
	}
	return regions, nil
}
