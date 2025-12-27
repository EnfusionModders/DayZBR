package postprocessing

import (
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
	"go.mongodb.org/mongo-driver/mongo"
)

func getAverageRating(db *database.BattleRoyaleDB, data []database.BRRawMatchPlayer) (float64, error) {
	count := 0.0
	average := 0.0
	for _, player := range data {
		rating := 0.0
		ldb_player, err := db.GetLeaderboardPlayerBySteamID(player.SteamID)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return 0, err
			}
			// player does not exist! assume default rating
			rating = DEFAULT_RATING
		} else {
			rating = ldb_player.Rating
		}
		if average == 0.0 {
			average = rating
			count++
		} else {
			//calculate weighted average
			average = average + ((rating - average) / count)
		}

	}
	return average, nil
}
