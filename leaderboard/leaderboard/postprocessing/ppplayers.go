package postprocessing

import (
	"math"

	elogo "github.com/kortemy/elo-go"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/database/database"
	"gitlab.desolationredux.com/DayZ/DayZBR-Mod/Service/leaderboard/leaderboard"
	"go.mongodb.org/mongo-driver/mongo"
)

func postProcessPlayers(ldb *leaderboard.Leaderboard, match *database.BRRawMatch, ldb_match *database.LeaderboardMatch) ([]*database.LeaderboardPlayer, error) {
	db := ldb.Db

	global_data, err := db.GetLeaderboardGlobaldata()
	if err != nil {
		ldb.Log.WithError(err).Errorln("failed to aquire global data. {error}")
		return nil, err
	}

	processed_players := make([]*database.LeaderboardPlayer, 0)

	average_rating, err := getAverageRating(ldb.Db, match.Results)
	if err != nil {
		ldb.Log.WithError(err).Errorln("first kill error. {error}")
		average_rating = DEFAULT_RATING
	}

	total_players := len(match.Results)

	total_distance_walked := float64(0)
	total_shots_fired := 0
	total_distance_driven := float64(0)
	total_items_pickedup := 0
	total_time_alive := int64(0)

	for i, player := range match.Results {

		raw_player, err := db.GetPlayerBySteamID(player.SteamID)
		if err != nil {
			ldb.Log.WithError(err).WithField("steamid", player.SteamID).Errorln("failed to aquire player. {steamid}. {error}")
			return nil, err
		}

		ldb_player, err := db.GetLeaderboardPlayerBySteamID(player.SteamID)
		if err == mongo.ErrNoDocuments {
			ldb_player, err = db.InsertLeaderboardPlayer(&database.LeaderboardPlayer{
				RawPlayerId: raw_player.Id,
				SteamID:     player.SteamID,
				Description: "",
				Rating:      DEFAULT_RATING,
				Pentagon: database.LeaderboardPlayerPentagon{
					Looting:   50,
					Fighting:  50,
					Walking:   50,
					Surviving: 50,
					Driving:   50,
				},
				RatingOverTime:         make(map[int64]float64),
				AveragePlacement:       1,
				TotalKills:             0,
				TotalWins:              0,
				TotalMatches:           0,
				TotalLooted:            0,
				TotalHits:              0,
				TotalDistanceOnFoot:    0,
				TotalDistanceInVehicle: 0,
				TotalTimeAlive:         0,
				MaxPossibleTimeAlive:   0,
				AverageTimeAlive:       0,
				Matches:                make([]database.LeaderboardPlayerMatch, 0),
			})
		}
		if err != nil {
			return nil, err
		}

		//1. update rating
		old_rating := ldb_player.Rating
		new_rating, err := calculateElo(old_rating, average_rating, total_players-(i+1), i)
		if err != nil {
			ldb.Log.WithError(err).Errorln("failed to calculate elo. {error}")
			return nil, err
		}
		ldb_player.Rating = new_rating

		//3. add new rating to rating over time
		ldb_player.RatingOverTime[match.Timestamp] = new_rating

		//4. update totals
		kills, err := getKills(ldb_player.SteamID, match.Results)
		if err != nil {
			ldb.Log.WithError(err).Errorln("failed to get kills. {error}")
			return nil, err
		}
		ldb_player.TotalKills += kills

		if i == len(match.Results)-1 {
			ldb_player.TotalWins++
		} else if i < 10 {
			ldb_player.TotalTop10++
		}

		ldb_player.TotalMatches++

		loot_pickups, err := getTotalLootPickups(ldb_player.SteamID, match.Events.Loots)
		if err != nil {
			ldb.Log.WithError(err).Errorln("failed to get loot pickups. {error}")
			return nil, err
		}
		ldb_player.TotalLooted += int64(loot_pickups)

		shots, err := getTotalShotsHittingTarget(ldb_player.SteamID, match.Events.Hits)
		if err != nil {
			ldb.Log.WithError(err).Errorln("failed to get total shots. {error}")
			return nil, err
		}
		ldb_player.TotalHits += int64(shots)

		distance_on_foot, err := getTotalDistanceOnFoot(ldb_player.SteamID, match.Events.Movements, match.Events.Vehicles)
		if err != nil {
			ldb.Log.WithError(err).Errorln("failed to get dist on foot. {error}")
			return nil, err
		}
		ldb_player.TotalDistanceOnFoot += distance_on_foot

		distance_in_vehicle, err := getTotalDistanceInVehicle(ldb_player.SteamID, match.Events.Movements, match.Events.Vehicles)
		if err != nil {
			ldb.Log.WithError(err).Errorln("failed to get dist on car. {error}")
			return nil, err
		}
		ldb_player.TotalDistanceInVehicle += distance_in_vehicle

		time_alive := int64(0)
		if player.DeathTimestamp == 0 {
			time_alive = match.Game.EndTime - match.Game.StartTime
		} else {
			time_alive = player.DeathTimestamp - match.Game.StartTime
		}

		ldb_player.TotalTimeAlive += time_alive
		ldb_player.MaxPossibleTimeAlive += match.Game.EndTime - match.Game.StartTime

		placement := total_players - i

		//append match info on matches
		ldb_player.Matches = append(ldb_player.Matches, database.LeaderboardPlayerMatch{
			MatchID:         ldb_match.Id,
			MatchName:       match.Game.MatchName,
			Placement:       placement,
			Kills:           kills,
			PostMatchRating: new_rating,
			DeltaRating:     new_rating - old_rating,
			Timestamp:       match.Timestamp,
		})

		// calculate averages
		if ldb_player.TotalMatches == 1 {
			//this is their first match!
			ldb_player.AveragePlacement = float64(placement)
			ldb_player.AverageTimeAlive = time_alive
		} else {
			ldb_player.AveragePlacement = ldb_player.AveragePlacement + ((float64(placement) - ldb_player.AveragePlacement) / float64(ldb_player.TotalMatches-1))
			ldb_player.AverageTimeAlive = ldb_player.AverageTimeAlive + ((time_alive - ldb_player.AverageTimeAlive) / int64(ldb_player.TotalMatches-1))
		}

		processed_players = append(processed_players, ldb_player)

		total_distance_walked += distance_on_foot
		total_distance_driven += distance_in_vehicle
		total_shots_fired += shots
		total_time_alive += time_alive
		total_items_pickedup += loot_pickups

	}

	average_distance_walked := total_distance_walked / float64(total_players)
	average_distance_driven := total_distance_driven / float64(total_players)
	average_shots_fired := float64(total_shots_fired) / float64(total_players)
	average_time_alive := float64(total_time_alive) / float64(total_players)
	average_loot_pickedup := float64(total_items_pickedup) / float64(total_players)

	//update globals!
	if global_data.TotalMatchesPlayed == 0 {

		global_data.PentagonData.AverageDistanceDrivenPerMatch = average_distance_driven
		global_data.PentagonData.AverageDistanceWalkedPerMatch = average_distance_walked
		global_data.PentagonData.AverageLootPerMatch = average_loot_pickedup
		global_data.PentagonData.AverageShotsPerMatch = average_shots_fired
		global_data.PentagonData.AveragePercentTimeAlivePerMatch = average_time_alive

	} else {

		global_data.PentagonData.AverageDistanceDrivenPerMatch = global_data.PentagonData.AverageDistanceDrivenPerMatch + (average_distance_driven-global_data.PentagonData.AverageDistanceDrivenPerMatch)/float64(global_data.TotalMatchesPlayed)
		global_data.PentagonData.AverageDistanceWalkedPerMatch = global_data.PentagonData.AverageDistanceWalkedPerMatch + (average_distance_walked-global_data.PentagonData.AverageDistanceWalkedPerMatch)/float64(global_data.TotalMatchesPlayed)
		global_data.PentagonData.AverageLootPerMatch = global_data.PentagonData.AverageLootPerMatch + (average_loot_pickedup-global_data.PentagonData.AverageLootPerMatch)/float64(global_data.TotalMatchesPlayed)
		global_data.PentagonData.AverageShotsPerMatch = global_data.PentagonData.AverageShotsPerMatch + (average_shots_fired-global_data.PentagonData.AverageShotsPerMatch)/float64(global_data.TotalMatchesPlayed)
		global_data.PentagonData.AveragePercentTimeAlivePerMatch = global_data.PentagonData.AveragePercentTimeAlivePerMatch + (average_time_alive-global_data.PentagonData.AveragePercentTimeAlivePerMatch)/float64(global_data.TotalMatchesPlayed)

	}
	global_data.TotalMatchesPlayed++
	global_data.TotalPlayers += len(match.Results)
	global_data.TotalDeadPlayers += global_data.TotalPlayers - 1
	global_data.ZombieKills += len(match.Events.ZombieKills)

	hex := global_data.Id.Hex()
	global_data, err = db.UpdateLeaderboardGlobalData(global_data)
	if err != nil {
		ldb.Log.WithError(err).WithField("hex", hex).Errorln("failed to update global data. {hex}. {error}")
		return nil, err
	}

	//update pentagon values! & push players to DB

	for i, player := range processed_players {

		loot_player_avg := float64(player.TotalLooted) / float64(player.TotalMatches)
		walked_player_avg := float64(player.TotalDistanceOnFoot) / float64(player.TotalMatches)
		driven_player_avg := float64(player.TotalDistanceInVehicle) / float64(player.TotalMatches)
		alive_player_avg := float64(player.AverageTimeAlive) / float64(player.MaxPossibleTimeAlive) //% time alive on average
		shots_player_avg := float64(player.TotalHits) / float64(player.TotalMatches)

		loot_global_avg := global_data.PentagonData.AverageLootPerMatch
		walked_global_avg := global_data.PentagonData.AverageDistanceWalkedPerMatch
		driven_global_avg := global_data.PentagonData.AverageDistanceDrivenPerMatch
		alive_global_avg := global_data.PentagonData.AveragePercentTimeAlivePerMatch
		shots_global_avg := global_data.PentagonData.AverageShotsPerMatch

		if loot_global_avg == 0 || driven_global_avg == 0 || shots_global_avg == 0 || alive_global_avg == 0 || walked_global_avg == 0 {
			processed_players[i].Pentagon.Looting = 50
			processed_players[i].Pentagon.Driving = 50
			processed_players[i].Pentagon.Fighting = 50
			processed_players[i].Pentagon.Surviving = 50
			processed_players[i].Pentagon.Walking = 50
		} else {
			//y=((tanh((((x)/2)-0.5)*3.14) + 1) / 2)
			processed_players[i].Pentagon.Looting = ((math.Tanh((((loot_player_avg/loot_global_avg)/2)-0.5)*math.Pi) + 1) / 2) * 100
			processed_players[i].Pentagon.Driving = ((math.Tanh((((driven_player_avg/driven_global_avg)/2)-0.5)*math.Pi) + 1) / 2) * 100
			processed_players[i].Pentagon.Fighting = ((math.Tanh((((shots_player_avg/shots_global_avg)/2)-0.5)*math.Pi) + 1) / 2) * 100
			processed_players[i].Pentagon.Surviving = ((math.Tanh((((alive_player_avg/alive_global_avg)/2)-0.5)*math.Pi) + 1) / 2) * 100
			processed_players[i].Pentagon.Walking = ((math.Tanh((((walked_player_avg/walked_global_avg)/2)-0.5)*math.Pi) + 1) / 2) * 100
		}

		_, err := db.UpdateLeaderboardPlayer(processed_players[i])
		if err != nil {
			ldb.Log.WithError(err).Errorln("failed to update player. {error}")
			return nil, err
		}

	}

	return processed_players, nil
}

type PositionedEvent struct {
	Player    string
	Position  []float64
	Timestamp int64
	EventType int
}

func getTotalDistanceInVehicle(steamid string, events []database.BRRawMatchMovementEvent, events2 []database.BRRawMatchVehicleEvent) (float64, error) {
	pos_events := make([]PositionedEvent, 0)

	for _, v_event := range events2 {
		if v_event.Player == steamid {
			pos_events = append(pos_events, PositionedEvent{
				Player:    v_event.Player,
				Position:  v_event.Position,
				Timestamp: v_event.Timestamp,
				EventType: v_event.Event,
			})
		}
	}
	for _, m_event := range events {
		if m_event.Player == steamid {
			pos_events = append(pos_events, PositionedEvent{
				EventType: -1,
				Player:    m_event.Player,
				Position:  m_event.Position,
				Timestamp: m_event.Timestamp,
			})
		}
	}

	distance_traveled := 0.0
	is_on_foot := true

	previous_event := PositionedEvent{
		Player: "",
	}
	for _, event := range pos_events {

		if previous_event.Player != "" { //check if we've given previous event a value

			if is_on_foot {

				if event.EventType == database.Event_GetIn {
					is_on_foot = false
					previous_event = event
				}

			} else {

				distance := math.Sqrt(
					math.Pow(previous_event.Position[0]-event.Position[0], 2) +
						math.Pow(previous_event.Position[1]-event.Position[1], 2) +
						math.Pow(previous_event.Position[2]-event.Position[2], 2))

				distance_traveled += distance

				if event.EventType == database.Event_GetOut {
					is_on_foot = true
				}

			}

		} else {
			previous_event = event
		}
	}

	return distance_traveled, nil
}

func getTotalDistanceOnFoot(steamid string, events []database.BRRawMatchMovementEvent, events2 []database.BRRawMatchVehicleEvent) (float64, error) {

	pos_events := make([]PositionedEvent, 0)

	for _, v_event := range events2 {
		if v_event.Player == steamid {
			pos_events = append(pos_events, PositionedEvent{
				Player:    v_event.Player,
				Position:  v_event.Position,
				Timestamp: v_event.Timestamp,
				EventType: v_event.Event,
			})
		}
	}
	for _, m_event := range events {
		if m_event.Player == steamid {
			pos_events = append(pos_events, PositionedEvent{
				EventType: -1,
				Player:    m_event.Player,
				Position:  m_event.Position,
				Timestamp: m_event.Timestamp,
			})
		}
	}

	distance_traveled := 0.0
	is_on_foot := true

	previous_event := PositionedEvent{
		Player: "",
	}
	for _, event := range pos_events {

		if previous_event.Player != "" { //check if we've given previous event a value
			if is_on_foot {
				distance := math.Sqrt(
					math.Pow(previous_event.Position[0]-event.Position[0], 2) +
						math.Pow(previous_event.Position[1]-event.Position[1], 2) +
						math.Pow(previous_event.Position[2]-event.Position[2], 2))

				distance_traveled += distance

				if event.EventType == database.Event_GetIn {
					is_on_foot = false
				}

			} else {
				if event.EventType == database.Event_GetOut {
					is_on_foot = true
					previous_event = event
				}
			}
		} else {
			previous_event = event
		}
	}

	return distance_traveled, nil
}

func getTotalShotsHittingTarget(steamid string, events []database.BRRawMatchHitEvent) (int, error) {
	shots := 0
	for _, event := range events {
		if event.Shooter == steamid {
			shots++
		}
	}
	return shots, nil
}
func getTotalLootPickups(steamid string, events []database.BRRawMatchLootEvent) (int, error) {
	pickups := 0
	for _, event := range events {
		if event.Player == steamid {
			if event.Event == database.Event_PickUp {
				pickups++
			}
		}
	}
	return pickups, nil
}
func getKills(steamid string, deathlist []database.BRRawMatchPlayer) (int, error) {

	kills := 0
	for _, player := range deathlist {
		if player.SteamID != steamid {
			if player.KillerID == steamid {
				kills++
			}
		}
	}
	return kills, nil
}

func calculateElo(initial_rating float64, average_rating float64, num_players_lost int, num_players_beat int) (float64, error) {
	//we may want to change our elo algorithm
	elo := elogo.NewElo()

	rankA := int(initial_rating)
	rankB := int(average_rating)

	for i := 0; i < num_players_beat; i++ {
		rankA = elo.Rating(rankA, rankB, 1)
	}
	for i := 0; i < num_players_lost; i++ {
		rankA = elo.Rating(rankA, rankB, 0)
	}

	return float64(rankA), nil
}
