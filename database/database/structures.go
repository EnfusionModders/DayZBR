package database

import "go.mongodb.org/mongo-driver/bson/primitive"

type BRServer struct {
	Id          primitive.ObjectID   `bson:"_id" json:"_id"`
	Name        string               `bson:"name" json:"name"`
	Connection  string               `bson:"connection" json:"connection"`
	QueryPort   string               `bson:"query_port" json:"query_port"`
	Version     string               `bson:"version" json:"version"`
	LastStarted int64                `bson:"last_started" json:"last_started"`
	Matches     []primitive.ObjectID `bson:"matches" json:"matches"`
	Region      string               `bson:"region" json:"region"`
	Locked      uint8                `bson:"locked" json:"locked"`
}
type BRPlayer struct {
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	Name      string             `bson:"name" json:"name"`
	SteamId   string             `bson:"steam_id" json:"steam_id"`
	Ips       []string           `bson:"ips" json:"ips"`
	Purchases []string           `bson:"shop_purchases" json:"shop_purchases"`
}

type BRRawMatch struct {
	Id        primitive.ObjectID `bson:"_id" json:"_id"`
	ServerId  primitive.ObjectID `bson:"server" json:"server"`
	Weather   BRRawMatchWeather  `bson:"weather" json:"weather"`
	Game      BRRawMatchGame     `bson:"game" json:"game"`
	Results   []BRRawMatchPlayer `bson:"results" json:"results"`
	Events    BRRawMatchEvents   `bson:"events" json:"events"`
	Timestamp int64              `bson:"timestamp" json:"timestamp"`
}
type BRRawMatchWeather struct {
	Fog    float64 `bson:"fog" json:"fog"`
	Rain   float64 `bson:"rain" json:"rain"`
	Hour   uint    `bson:"hour" json:"hour"`
	Minute uint    `bson:"minute" json:"minute"`
}
type BRRawMatchGame struct {
	MatchName string `bson:"name" json:"name"`
	MapName   string `bson:"map" json:"mapname"`
	GameType  string `bson:"gametype" json:"gametype"`
	StartTime int64  `bson:"start" json:"start"`
	EndTime   int64  `bson:"end" json:"end"`
}
type BRRawMatchPlayer struct {
	SteamID        string    `bson:"steamid" json:"steamid"`
	KillerID       string    `bson:"killedby" json:"killedby"`
	KillerWeapon   string    `bson:"killedwith" json:"killedwith"`
	DeathTimestamp int64     `bson:"timestamp" json:"timestamp"`
	Position       []float64 `bson:"pos" json:"pos"`
	KillerPosition []float64 `bson:"killerpos" json:"killerpos"`
}

const (
	EventType_Airdrop int = iota
	EventType_CircleShow
	EventType_CircleLock
	EventType_Hit
	EventType_LootPickUp
	EventType_LootDrop
	EventType_Movement
	EventType_Shot
	EventType_VehicleGetIn
	EventType_VehicleGetOut
	EventType_ZombieKill
)

const (
	Event_GetIn int = iota
	Event_GetOut
)
const (
	Event_PickUp int = iota
	Event_Drop
)
const (
	Event_ShowCircle int = iota
	Event_LockCircle
)

type BRRawMatchEvents struct {
	ZombieKills []BRRawMatchZombieEvent   `bson:"zombiekills" json:"zombiekills"`
	Vehicles    []BRRawMatchVehicleEvent  `bson:"vehicles" json:"vehicles"`
	Shots       []BRRawMatchShotEvent     `bson:"shots" json:"shots"`
	Hits        []BRRawMatchHitEvent      `bson:"hits" json:"hits"`
	Movements   []BRRawMatchMovementEvent `bson:"movements" json:"movements"`
	Loots       []BRRawMatchLootEvent     `bson:"loots" json:"loots"`
	Circles     []BRRawMatchCircleEvent   `bson:"circles" json:"circles"`
	Airdrops    []BRRawMatchAirdropEvent  `bson:"airdrops" json:"airdrops"`
}
type BRRawMatchZombieEvent struct {
	Player    string    `bson:"playerid" json:"playerid"`
	Position  []float64 `bson:"pos" json:"pos"`
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
}
type BRRawMatchVehicleEvent struct {
	Player    string    `bson:"playerid" json:"playerid"`
	Vehicle   string    `bson:"vehicle" json:"vehicle"`
	Position  []float64 `bson:"pos" json:"pos"`
	Event     int       `bson:"event" json:"brevent"`
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
}
type BRRawMatchLootEvent struct {
	Player    string    `bson:"playerid" json:"playerid"`
	Item      string    `bson:"item" json:"item"`
	Position  []float64 `bson:"pos" json:"pos"`
	Event     int       `bson:"event" json:"brevent"`
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
}
type BRRawMatchShotEvent struct {
	Player    string    `bson:"playerid" json:"playerid"`
	Position  []float64 `bson:"pos" json:"pos"`
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
}
type BRRawMatchHitEvent struct {
	Player    string    `bson:"playerid" json:"playerid"`
	Position  []float64 `bson:"pos" json:"pos"`
	Shooter   string    `bson:"shooterid" json:"shooterid"`
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
}
type BRRawMatchMovementEvent struct {
	Player    string    `bson:"playerid" json:"playerid"`
	Position  []float64 `bson:"pos" json:"pos"`
	Direction float64   `bson:"direction" json:"direction"`
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
}
type BRRawMatchCircleEvent struct {
	Position  []float64 `bson:"pos" json:"pos"`
	Radius    float64   `bson:"radius" json:"radius"`
	Event     int       `bson:"event" json:"brevent"`
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
}
type BRRawMatchAirdropEvent struct {
	Position  []float64 `bson:"pos" json:"pos"`
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
}

type LeaderboardMatch struct {
	Id            primitive.ObjectID          `bson:"_id" json:"_id"`
	RawDataId     primitive.ObjectID          `bson:"rawmatch" json:"rawmatch"`
	Name          string                      `bson:"match_name" json:"match_name"`
	Winner        string                      `bson:"winner" json:"winner"`
	MostKills     string                      `bson:"most_kills" json:"most_kills"`
	FirstKill     string                      `bson:"first_kill" json:"first_kill"`
	LongestKill   string                      `bson:"longest_kill" json:"longest_kill"`
	AverageRating float64                     `bson:"average_rating" json:"average_rating"`
	MapName       string                      `bson:"map" json:"mapname"`
	GameType      string                      `bson:"gametype" json:"gametype"`
	PlayerCount   int                         `bson:"playercount" json:"playercount"`
	Duration      int64                       `bson:"duration" json:"duration"` //duration in ms/ns
	Weather       string                      `bson:"weather" json:"weather"`
	Events        []LeaderboardMatchEvent     `bson:"events" json:"events"`
	Placements    []LeaderboardMatchPlacement `bson:"placements" json:"placements"`
	Timestamp     int64                       `bson:"timestamp" json:"timestamp"`
}
type LeaderboardMatchEvent struct {
	Timestamp int64     `bson:"timestamp" json:"timestamp"`
	Position  []float64 `bson:"position" json:"position"`
	EventType int       `bson:"type" json:"type"`
	Data      string    `bson:"data" json:"data"`
}
type LeaderboardMatchPlacement struct {
	SteamID      string  `bson:"steamid" json:"steamid"`
	Kills        int     `bson:"kills" json:"kills"`
	TimeAlive    int64   `bson:"alive" json:"alive"`
	KilledBy     string  `bson:"killer" json:"killer"`
	KilledWith   string  `bson:"weapon" json:"weapon"`
	KillDistance float64 `bson:"distance" json:"distance"`
}

type LeaderboardPlayer struct {
	Id                     primitive.ObjectID        `bson:"_id" json:"_id"`
	RawPlayerId            primitive.ObjectID        `bson:"playerid" json:"playerid"`
	SteamID                string                    `bson:"steamid" json:"steamid"`
	Description            string                    `bson:"description" json:"description"`
	Rating                 float64                   `bson:"rating" json:"rating"`
	Pentagon               LeaderboardPlayerPentagon `bson:"pentagon" json:"pentagon"`
	RatingOverTime         map[int64]float64         `bson:"ratingtime" json:"ratingtime"`
	AveragePlacement       float64                   `bson:"averageplace" json:"averageplace"`
	TotalKills             int                       `bson:"totalkills" json:"totalkills"`
	TotalWins              int                       `bson:"totalwins" json:"totalwins"`
	TotalTop10             int                       `bson:"totaltop10" json:"totaltop10"`
	TotalMatches           int                       `bson:"totalmatches" json:"totalmatches"`
	TotalLooted            int64                     `bson:"totallooted" json:"totallooted"`
	TotalHits              int64                     `bson:"totalhits" json:"totalhits"`
	TotalDistanceOnFoot    float64                   `bson:"totaldistanceonfoot" json:"totaldistanceonfoot"`
	TotalDistanceInVehicle float64                   `bson:"totaldistanceinvehicle" json:"totaldistanceinvehicle"`
	TotalTimeAlive         int64                     `bson:"totaltimealive" json:"totaltimealive"`
	MaxPossibleTimeAlive   int64                     `bson:"maxpossibletimealive" json:"maxpossibletimealive"`
	AverageTimeAlive       int64                     `bson:"averagetimealive" json:"averagetimealive"`
	Matches                []LeaderboardPlayerMatch  `bson:"matches" json:"matches"`
}
type LeaderboardPlayerPentagon struct {
	Looting   float64 `bson:"looting" json:"looting"`
	Fighting  float64 `bson:"fighting" json:"fighting"`
	Walking   float64 `bson:"walking" json:"walking"`
	Surviving float64 `bson:"surviving" json:"surviving"`
	Driving   float64 `bson:"driving" json:"driving"`
}
type LeaderboardPlayerMatch struct {
	MatchID         primitive.ObjectID `bson:"matchid" json:"matchid"`
	MatchName       string             `bson:"matchname" json:"matchname"`
	Placement       int                `bson:"placement" json:"placement"`
	Kills           int                `bson:"kills" json:"kills"`
	PostMatchRating float64            `bson:"postmatchrating" json:"postmatchrating"`
	DeltaRating     float64            `bson:"deltarating" json:"deltarating"`
	Timestamp       int64              `bson:"timestamp" json:"timestamp"`
}

type GlobalLeaderboardData struct {
	Id                 primitive.ObjectID        `bson:"_id" json:"_id"`
	TotalMatchesPlayed int                       `bson:"totalmatches" json:"totalmatches"`
	TotalDeadPlayers   int                       `bson:"totaldead" json:"totaldead"`
	TotalPlayers       int                       `bson:"totalplayers" json:"totalplayers"`
	ZombieKills        int                       `bson:"zombiekills" json:"zombiekills"`
	PentagonData       GlobalAveragePentagonData `bson:"pentagon" json:"pentagon"`
}
type GlobalAveragePentagonData struct {
	AverageLootPerMatch             float64 `bson:"avgloot" json:"avgloot"`
	AverageShotsPerMatch            float64 `bson:"avgshots" json:"avgshots"`
	AverageDistanceWalkedPerMatch   float64 `bson:"avgdistancewalked" json:"avgdistancewalked"`
	AveragePercentTimeAlivePerMatch float64 `bson:"avgtimealive" json:"avgtimealive"`
	AverageDistanceDrivenPerMatch   float64 `bson:"avgdriven" json:"avgdriven"`
}
