package database

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type BattleRoyaleDB struct {
	Client    *mongo.Client
	Connected bool
}

func ConnectTo(connection_uri string) (*BattleRoyaleDB, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connection_uri))
	if err != nil {
		return nil, err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	db := &BattleRoyaleDB{
		Client:    client,
		Connected: true,
	}
	return db, nil
}

//--- functionality
func (db *BattleRoyaleDB) Disconnect() error {
	db.Connected = false
	if !db.Connected {
		return errors.New("not connected to database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := db.Client.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}
func (db *BattleRoyaleDB) GetCollection(collection string) *mongo.Collection {
	return db.Client.Database("dayzbr").Collection(collection)
}

func (db *BattleRoyaleDB) InsertRawMatch(match *BRRawMatch) (*BRRawMatch, error) {
	match.Timestamp = int64(time.Now().Unix())
	match.Id = primitive.NewObjectID()

	collection := db.GetCollection("raw_match_data")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, match)
	return match, err
}

func (db *BattleRoyaleDB) GetRawMatchByID(matchid string) (*BRRawMatch, error) {
	collection := db.GetCollection("raw_match_data")

	match_id, err := primitive.ObjectIDFromHex(matchid)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"_id": match_id})
	err = res.Err()
	if err != nil {
		return nil, err
	}
	var match BRRawMatch
	err = res.Decode(&match)
	return &match, err
}
func (db *BattleRoyaleDB) GetAllServers() ([]*BRServer, error) {
	collection := db.GetCollection("servers")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	it, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	var results []BRServer
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = it.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	var response []*BRServer
	for i := range results {
		response = append(response, &results[i])
	}
	return response, nil
}
func (db *BattleRoyaleDB) GetServersFiltered(filter interface{}) ([]*BRServer, error) {
	collection := db.GetCollection("servers")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	it, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var results []BRServer
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = it.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	var response []*BRServer
	for i := range results {
		response = append(response, &results[i])
	}
	return response, nil
}
func (db *BattleRoyaleDB) GetServersInRegion(version string, regioncode string) ([]*BRServer, error) {
	collection := db.GetCollection("servers")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	it, err := collection.Find(ctx, bson.M{
		"locked":  0,
		"version": version,
		"region":  bson.M{"$in": []string{"any", regioncode}},
	})
	if err != nil {
		return nil, err
	}
	var results []BRServer
	err = it.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	var response []*BRServer
	for i := range results {
		response = append(response, &results[i])
	}
	return response, nil
}
func (db *BattleRoyaleDB) InsertServer(server *BRServer) (*BRServer, error) {
	server.LastStarted = int64(time.Now().Unix())
	server.Id = primitive.NewObjectID()
	collection := db.GetCollection("servers")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, server)
	return server, err
}
func (db *BattleRoyaleDB) GetServerByConnection(ip string, queryport string) (*BRServer, error) {
	collection := db.GetCollection("servers")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{
		"connection": bson.M{"$regex": ip + ".*"},
		"query_port": queryport,
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}
	var server BRServer
	err = res.Decode(&server)
	return &server, err
}
func (db *BattleRoyaleDB) UpdateServer(server *BRServer) (*BRServer, error) {
	collection := db.GetCollection("servers")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOneAndReplace(ctx, bson.M{"_id": server.Id}, server)
	return server, res.Err()
}
func (db *BattleRoyaleDB) GetServerByID(serverid string) (*BRServer, error) {
	collection := db.GetCollection("servers")

	server_id, err := primitive.ObjectIDFromHex(serverid)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"_id": server_id})
	err = res.Err()
	if err != nil {
		return nil, err
	}
	var server BRServer
	err = res.Decode(&server)
	return &server, err
}

func (db *BattleRoyaleDB) GetPlayerByID(playerid string) (*BRPlayer, error) {
	collection := db.GetCollection("players")

	player_id, err := primitive.ObjectIDFromHex(playerid)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"_id": player_id})
	err = res.Err()
	if err != nil {
		return nil, err
	}
	var player BRPlayer
	err = res.Decode(&player)
	return &player, err
}
func (db *BattleRoyaleDB) GetPlayerBySteamID(steamid string) (*BRPlayer, error) {
	collection := db.GetCollection("players")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{
		"steam_id": steamid,
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}
	var player BRPlayer
	err = res.Decode(&player)
	return &player, err
}
func (db *BattleRoyaleDB) UpdatePlayer(player *BRPlayer) (*BRPlayer, error) {
	collection := db.GetCollection("players")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOneAndReplace(ctx, bson.M{"_id": player.Id}, player)
	return player, res.Err()
}
func (db *BattleRoyaleDB) InsertPlayer(player *BRPlayer) (*BRPlayer, error) {
	player.Id = primitive.NewObjectID()
	collection := db.GetCollection("players")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, player)
	return player, err
}

//--- new leaderboard functions
func (db *BattleRoyaleDB) UpdateLeaderboardPlayer(player *LeaderboardPlayer) (*LeaderboardPlayer, error) {
	collection := db.GetCollection("leaderboard_players")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOneAndReplace(ctx, bson.M{"_id": player.Id}, player)
	return player, res.Err()
}
func (db *BattleRoyaleDB) InsertLeaderboardPlayer(player *LeaderboardPlayer) (*LeaderboardPlayer, error) {
	player.Id = primitive.NewObjectID()
	collection := db.GetCollection("leaderboard_players")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, player)
	return player, err
}
func (db *BattleRoyaleDB) GetLeaderboardPlayerBySteamID(steamid string) (*LeaderboardPlayer, error) {
	collection := db.GetCollection("leaderboard_players")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{
		"steamid": steamid,
	})
	err := res.Err()
	if err != nil {
		return nil, err
	}
	var player LeaderboardPlayer
	err = res.Decode(&player)
	return &player, err
}
func (db *BattleRoyaleDB) GetLeaderboardPlayerByID(playerid string) (*LeaderboardPlayer, error) {
	collection := db.GetCollection("leaderboard_players")

	player_id, err := primitive.ObjectIDFromHex(playerid)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"_id": player_id})
	err = res.Err()
	if err != nil {
		return nil, err
	}
	var player LeaderboardPlayer
	err = res.Decode(&player)
	return &player, err
}
func (db *BattleRoyaleDB) GetLeaderboardPlayerByPlayerID(playerid string) (*LeaderboardPlayer, error) {
	collection := db.GetCollection("leaderboard_players")

	player_id, err := primitive.ObjectIDFromHex(playerid)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"playerid": player_id})
	err = res.Err()
	if err != nil {
		return nil, err
	}
	var player LeaderboardPlayer
	err = res.Decode(&player)
	return &player, err
}
func (db *BattleRoyaleDB) GetLeaderboardRank(rating float64) (int64, error) {
	//counts the number of elements in leaderbaord with rating > provided value

	collection := db.GetCollection("leaderboard_players")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//get count of documents ahead of me (that determines my rank)
	filter := bson.M{
		"rating": bson.M{"$gt": rating},
	}
	res, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return -1, err
	}
	return res + 1, nil //we want to return 1 if 0 players have a higher rating
}
func (db *BattleRoyaleDB) GetTopLeaderboardPlayersByRating(player_limit int64) ([]LeaderboardPlayer, error) {

	collection := db.GetCollection("leaderboard_players")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	options := options.Find()
	options.SetSort(map[string]int{"rating": -1}) //rating decending
	options.SetLimit(player_limit)

	it, err := collection.Find(ctx, bson.D{}, options)
	if err != nil {
		return nil, err
	}
	var results []LeaderboardPlayer
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = it.All(ctx, &results)
	return results, err
}

func (db *BattleRoyaleDB) InsertLeaderboardMatch(match *LeaderboardMatch) (*LeaderboardMatch, error) {
	match.Id = primitive.NewObjectID()
	collection := db.GetCollection("leaderboard_matches")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, match)
	return match, err
}
func (db *BattleRoyaleDB) GetLeaderboardMatchByID(matchid string) (*LeaderboardMatch, error) {
	collection := db.GetCollection("leaderboard_matches")

	match_id, err := primitive.ObjectIDFromHex(matchid)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"_id": match_id})
	err = res.Err()
	if err != nil {
		return nil, err
	}
	var match LeaderboardMatch
	err = res.Decode(&match)
	return &match, err
}
func (db *BattleRoyaleDB) GetLeaderboardMatchByRawID(matchid string) (*LeaderboardMatch, error) {
	collection := db.GetCollection("leaderboard_matches")

	match_id, err := primitive.ObjectIDFromHex(matchid)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{"rawmatch": match_id})
	err = res.Err()
	if err != nil {
		return nil, err
	}
	var match LeaderboardMatch
	err = res.Decode(&match)
	return &match, err
}
func (db *BattleRoyaleDB) UpdateLeaderboardGlobalData(data *GlobalLeaderboardData) (*GlobalLeaderboardData, error) {
	collection := db.GetCollection("leaderboard_globals")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOneAndReplace(ctx, bson.M{"_id": data.Id}, data)
	return data, res.Err()
}
func (db *BattleRoyaleDB) GetLeaderboardGlobaldata() (*GlobalLeaderboardData, error) {
	collection := db.GetCollection("leaderboard_globals")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := collection.FindOne(ctx, bson.M{})
	err := res.Err()
	if err != nil {
		return nil, err
	}
	var data GlobalLeaderboardData
	err = res.Decode(&data)
	return &data, err
}
