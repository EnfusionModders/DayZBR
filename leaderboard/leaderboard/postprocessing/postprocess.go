package postprocessing

import (
	"github.com/sirupsen/logrus"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/leaderboard/leaderboard"
)

func ProcessNewMatchRoutine(ldb *leaderboard.Leaderboard) {

	for {
		match, more := <-ldb.Chn

		if !more {
			ldb.Log.Warnln("Process New Match Channel Closed!")
			return
		}

		//1. post process match
		ldb_match, err := postProcessMatch(ldb, match)
		if err != nil {
			continue //error already logged
		}
		//3. for each player, create or update their db object
		players, err := postProcessPlayers(ldb, match, ldb_match)
		if err != nil {
			ldb.Log.Error("failed to process players", err)
			continue
		}

		ldb.Log.WithFields(logrus.Fields{
			"match_id":     ldb_match.Id.Hex(),
			"player_count": len(players),
		}).Info("Leaderbaords Updated. Inserted new match {match_id} with {player_count} players.")
	}

}
