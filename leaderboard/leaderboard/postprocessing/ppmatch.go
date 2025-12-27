package postprocessing

import (
	"fmt"
	"math"
	"sort"

	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/leaderboard/leaderboard"
)

const DEFAULT_RATING float64 = 1000.0

func postProcessMatch(ldb *leaderboard.Leaderboard, match *database.BRRawMatch) (*database.LeaderboardMatch, error) {
	most_kills, err := getMostKills(match.Results)
	if err != nil {
		ldb.Log.WithError(err).Errorln("most kills error. {error}")
		most_kills = "00000000000000000"
	}
	first_kill, err := getFirstKill(match.Results)
	if err != nil {
		ldb.Log.WithError(err).Errorln("first kill error. {error}")
		first_kill = "00000000000000000"
	}
	longest_kill, err := getLongestKill(match.Results)
	if err != nil {
		ldb.Log.WithError(err).Errorln("longest kill error. {error}")
		longest_kill = "00000000000000000"
	}
	rating_avg, err := getAverageRating(ldb.Db, match.Results)
	if err != nil {
		ldb.Log.WithError(err).Errorln("average rating error. {error}")
		rating_avg = DEFAULT_RATING
	}
	weather, err := getWeatherString(match.Weather)
	if err != nil {
		ldb.Log.WithError(err).Errorln("weather error. {error}")
		weather = "Clear Day"
	}
	events, err := getMatchEvents(match, match.Events)
	if err != nil {
		ldb.Log.WithError(err).Errorln("events error. {error}")
		events = make([]database.LeaderboardMatchEvent, 0)
	}
	placements, err := getMatchPlacements(match, match.Results)
	if err != nil {
		ldb.Log.WithError(err).Errorln("placements error. {error}")
		placements = make([]database.LeaderboardMatchPlacement, 0)
	}

	ldb_match := &database.LeaderboardMatch{
		RawDataId:     match.Id,
		Name:          match.Game.MatchName,
		Winner:        match.Results[len(match.Results)-1].SteamID, //match.Results is reverse order (index 0 is first death)
		MostKills:     most_kills,
		FirstKill:     first_kill,
		LongestKill:   longest_kill,
		AverageRating: rating_avg,
		MapName:       match.Game.MapName,
		GameType:      match.Game.GameType,
		PlayerCount:   len(match.Results),
		Duration:      match.Game.EndTime - match.Game.StartTime,
		Weather:       weather,
		Events:        events,
		Placements:    placements,
		Timestamp:     match.Timestamp - match.Game.EndTime, //timestamp here marks time at start (whereas in raw it's time at end)
	}

	//2. create match db object
	ldb_match, err = ldb.Db.InsertLeaderboardMatch(ldb_match)
	if err != nil {
		ldb.Log.WithError(err).Errorln("failed to insert leaderboard match. {error}")
		return nil, err
	}

	return ldb_match, nil
}

func getMatchPlacements(match *database.BRRawMatch, data []database.BRRawMatchPlayer) ([]database.LeaderboardMatchPlacement, error) {
	places := make([]database.LeaderboardMatchPlacement, 0)

	//1. insert all of the players into the leaderbaord match placement list
	for _, player := range data {
		//we prepend the new placement, as we need to reverse the order from death data to placement
		places = append([]database.LeaderboardMatchPlacement{
			{
				SteamID:    player.SteamID,
				Kills:      0, //we'll update this as we add more players
				TimeAlive:  player.DeathTimestamp - match.Game.StartTime,
				KilledBy:   player.KillerID,
				KilledWith: player.KillerWeapon,
				KillDistance: math.Sqrt(
					math.Pow(player.Position[0]-player.KillerPosition[0], 2) +
						math.Pow(player.Position[1]-player.KillerPosition[1], 2) +
						math.Pow(player.Position[2]-player.KillerPosition[2], 2)),
			},
		}, places...)
	}

	for i, placement := range places {
		kills := 0
		for _, placement2 := range places {
			if placement.SteamID != placement2.SteamID {
				if placement.SteamID == placement2.KilledBy {
					kills++
				}
			}
		}
		places[i].Kills = kills //update value in the array
	}

	return places, nil
}
func getMatchEvents(match *database.BRRawMatch, data database.BRRawMatchEvents) ([]database.LeaderboardMatchEvent, error) {

	events := make([]database.LeaderboardMatchEvent, 0)

	for _, event := range data.Airdrops {
		events = append(events, database.LeaderboardMatchEvent{
			Timestamp: event.Timestamp - match.Game.StartTime,
			Position:  event.Position,
			EventType: database.EventType_Airdrop,
			Data:      "",
		})
	}
	for _, event := range data.Circles {
		events = append(events, database.LeaderboardMatchEvent{
			Timestamp: event.Timestamp - match.Game.StartTime,
			Position:  event.Position,
			EventType: database.EventType_CircleShow + event.Event,
			Data:      fmt.Sprint(event.Radius),
		})
	}
	for _, event := range data.Hits {
		events = append(events, database.LeaderboardMatchEvent{
			Timestamp: event.Timestamp - match.Game.StartTime,
			Position:  event.Position,
			EventType: database.EventType_Hit,
			Data:      event.Player + "," + event.Shooter,
		})
	}
	for _, event := range data.Loots {
		events = append(events, database.LeaderboardMatchEvent{
			Timestamp: event.Timestamp - match.Game.StartTime,
			Position:  event.Position,
			EventType: database.EventType_LootPickUp + event.Event,
			Data:      event.Player + "," + event.Item,
		})
	}
	for _, event := range data.Movements {
		events = append(events, database.LeaderboardMatchEvent{
			Timestamp: event.Timestamp - match.Game.StartTime,
			Position:  event.Position,
			EventType: database.EventType_Movement,
			Data:      event.Player + "," + fmt.Sprint(event.Direction),
		})
	}
	for _, event := range data.Shots {
		events = append(events, database.LeaderboardMatchEvent{
			Timestamp: event.Timestamp - match.Game.StartTime,
			Position:  event.Position,
			EventType: database.EventType_Shot,
			Data:      event.Player,
		})
	}
	for _, event := range data.Vehicles {
		events = append(events, database.LeaderboardMatchEvent{
			Timestamp: event.Timestamp - match.Game.StartTime,
			Position:  event.Position,
			EventType: database.EventType_VehicleGetIn + event.Event,
			Data:      event.Player + "," + event.Vehicle,
		})
	}
	for _, event := range data.ZombieKills {
		events = append(events, database.LeaderboardMatchEvent{
			Timestamp: event.Timestamp - match.Game.StartTime,
			Position:  event.Position,
			EventType: database.EventType_ZombieKill,
			Data:      event.Player,
		})
	}

	//sort events by timestamp
	sort.Slice(events[:], func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})

	return events, nil
}

func getWeatherString(weather database.BRRawMatchWeather) (string, error) {
	time := "Day"
	if weather.Hour > 20 || weather.Hour < 7 {
		time = "Night"
	}
	mood := "Clear"
	if weather.Fog > 0.2 {
		mood = "Foggy"
	}
	if weather.Rain > 0.2 {
		mood = "Rainy"
	}

	return mood + " " + time, nil
}

func getLongestKill(data []database.BRRawMatchPlayer) (string, error) {

	longest_kill := "00000000000000000"
	longest_kill_dist := 0.0
	for _, player := range data {
		if player.KillerID != "" && player.KillerID != player.SteamID {
			dist := math.Sqrt(
				math.Pow(player.Position[0]-player.KillerPosition[0], 2) +
					math.Pow(player.Position[1]-player.KillerPosition[1], 2) +
					math.Pow(player.Position[2]-player.KillerPosition[2], 2))
			if dist > longest_kill_dist {
				longest_kill = player.KillerID
				longest_kill_dist = dist
			}
		}
	}

	return longest_kill, nil
}

func getFirstKill(data []database.BRRawMatchPlayer) (string, error) {
	for _, player := range data {
		if player.KillerID != "" && player.SteamID != player.KillerID {
			return player.KillerID, nil
		}
	}

	return "00000000000000000", nil
}
func getMostKills(data []database.BRRawMatchPlayer) (string, error) {
	kills := make(map[string]int)
	for _, player := range data {
		if player.KillerID != "" && player.SteamID != player.KillerID {
			kills[player.KillerID]++
		}
	}
	most_kills := "00000000000000000"
	most_kills_count := 0

	for id, kc := range kills {
		if kc > most_kills_count {
			most_kills_count = kc
			most_kills = id
		}
	}
	return most_kills, nil
}
