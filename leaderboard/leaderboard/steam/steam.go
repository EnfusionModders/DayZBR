package steam

import (
	"errors"
	"strconv"

	"github.com/Philipp15b/go-steamapi"
)

func GetPlayerInfo(steamid string, apikey string) (*steamapi.PlayerSummary, error) {

	asInt, err := strconv.ParseUint(steamid, 10, 64)
	if err != nil {
		return nil, err
	}

	profiles, err := steamapi.GetPlayerSummaries([]uint64{asInt}, apikey)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, errors.New("summery not found. possibly invalid steamid")
	}

	return &profiles[0], nil
}
