package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/michelemendel/bhbe/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var lg *zap.SugaredLogger

type RedisCtx struct {
	rdb *redis.Client
}

// prefix f=field
const (
	fUUID         = "uuid"
	fName         = "name"
	fUpdatedAt    = "updatedAt"
	fGeoUpdatedAt = "geoUpdatedAt"
	locations     = "locations"
	ClientPrefix  = "client:"
)

type Client struct {
	UUID         string    `json:"uuid"`
	Name         string    `json:"name"`
	UpdatedAt    time.Time `json:"updatedAt"`
	GeoUpdatedAt time.Time `json:"geoUpdatedAt"`
}

type Geo struct {
	UUID    string  `json:"uuid"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	DistKm  float64 `json:"distKm"`
	GeoHash int64   `json:"geohash"`
}

type ClientGeo struct {
	Client Client `json:"client"`
	Geo    Geo    `json:"geo"`
}

func init() {
	lg = utils.Log()
}

func InitRedisClient() *RedisCtx {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_ADDRESS"), os.Getenv("REDIS_PORT")),
	})

	return &RedisCtx{rdb: rdb}
}

// --------------------------------------------------------------------------------
// General

func (r RedisCtx) GetKeys(pattern string) []string {
	ctx := context.Background()
	var vals []string
	iter := r.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		vals = append(vals, iter.Val())
	}
	return vals
}

// --------------------------------------------------------------------------------
// Client

func (r RedisCtx) setClient(uuid, name string) string {
	ctx := context.Background()

	if name == "" {
		name = "Anonymous"
	}

	err := r.rdb.HSet(ctx, uuid, fUUID, uuid, fName, name, fUpdatedAt, utils.Now()).Err()
	if err != nil {
		lg.Warnf("[redis] Error setting client: %s. Err: %s", uuid, err)
	}
	return uuid
}

func (r RedisCtx) updateClientGeoUpdateAt(uuid string) string {
	ctx := context.Background()

	err := r.rdb.HSet(ctx, uuid, fGeoUpdatedAt, utils.Now()).Err()
	if err != nil {
		lg.Warnf("[redis] Error setting client: %s. Err: %s", uuid, err)
	}
	return uuid
}

func (r RedisCtx) CreateClient(clientPrefix, name string) string {
	return r.setClient(clientPrefix+utils.GenerateUUID(), name)
}

func (r RedisCtx) UpsertClient(uuid, name string) string {
	return r.setClient(uuid, name)
}

func (r RedisCtx) DeleteClient(uuid string) string {
	ctx := context.Background()
	r.rdb.Del(ctx, uuid)
	return uuid
}

func (r RedisCtx) DeleteClients(uuids []string) []string {
	ctx := context.Background()
	r.rdb.Del(ctx, uuids...)
	return uuids
}

func (r RedisCtx) GetClient(uuid string) *Client {
	ctx := context.Background()

	// Scan can't scan time
	// err := r.rdb.HMGet(ctx, uuid, fName, fUpdatedAt).Scan(&client)
	client, err := r.rdb.HMGet(ctx, uuid, fUUID, fName, fUpdatedAt, fGeoUpdatedAt).Result()
	if err != nil {
		lg.Infof("[redis] Error getting UUID %s. Err: %s ", uuid, err)
	}

	// fUUID
	if client[0] == nil {
		client[0] = ""
	}
	// fName
	if client[1] == nil {
		client[1] = ""
	}
	// fUpdatedAt
	updatedAt, _ := time.Parse(time.RFC3339, client[2].(string))
	// fGeoUpdatedAt
	if client[3] == nil {
		client[3] = "0"
	}
	geoUpdatedAt, _ := time.Parse(time.RFC3339, client[3].(string))

	return &Client{
		UUID:         client[0].(string),
		Name:         client[1].(string),
		UpdatedAt:    updatedAt,
		GeoUpdatedAt: geoUpdatedAt,
	}
}

func (r RedisCtx) GetClients(pattern string) []Client {
	ctx := context.Background()
	var clients []Client
	iter := r.rdb.Scan(ctx, 0, ClientPrefix+pattern, 0).Iterator()
	for iter.Next(ctx) {
		fmt.Println("iter.Val(): ", iter.Val())
		clients = append(clients, *r.GetClient(iter.Val()))
	}
	return clients
}

// --------------------------------------------------------------------------------
// Geo

func (r RedisCtx) UpsertGeo(uuid string, lat, lng float64) string {
	ctx := context.Background()

	geoLocation := &redis.GeoLocation{
		Name:      uuid,
		Longitude: lng,
		Latitude:  lat,
	}

	_, err := r.rdb.GeoAdd(ctx, locations, geoLocation).Result()
	if err != nil {
		lg.Infof("[redis] Error setting geo for UUID %s. Err: %s ", uuid, err)
	}

	r.updateClientGeoUpdateAt(uuid)
	return uuid
}

func (r RedisCtx) GetGeo(uuid string) *Geo {
	ctx := context.Background()
	res, err := r.rdb.GeoPos(ctx, locations, uuid).Result()
	if err != nil {
		lg.Infof("[redis] Error getting geo for UUID %s. Err: %s ", uuid, err)
	}

	if res[0] == nil {
		return &Geo{
			UUID: uuid,
			Lat:  0,
			Lng:  0,
		}
	}

	return &Geo{
		UUID: uuid,
		Lat:  res[0].Latitude,
		Lng:  res[0].Longitude,
	}
}

func (r RedisCtx) DelGeo(uuid string) string {
	ctx := context.Background()
	_, err := r.rdb.ZRem(ctx, locations, uuid).Result()
	if err != nil {
		lg.Infof("[redis] Error deleting geo for UUID %s. Err: %s ", uuid, err)
	}
	return uuid
}

func (r RedisCtx) DelGeos(uuids []string) []string {
	ctx := context.Background()
	if uuids == nil {
		return []string{}
	}
	_, err := r.rdb.ZRem(ctx, locations, uuids).Result()
	if err != nil {
		lg.Infof("[redis] Error deleting geos. Err: %s ", uuids, err)
	}
	return uuids
}

func (r RedisCtx) SearchLocationQuery(lat, lng, radius float64) *redis.GeoSearchLocationQuery {
	return &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Latitude:   lat,
			Longitude:  lng,
			Radius:     radius,
			RadiusUnit: "km",
			Sort:       "asc",
		},
		WithHash:  true,
		WithDist:  true,
		WithCoord: true,
	}
}

func (r RedisCtx) SearchLocation(query *redis.GeoSearchLocationQuery) []Geo {
	ctx := context.Background()
	res, err := r.rdb.GeoSearchLocation(ctx, locations, query).Result()
	if err != nil {
		lg.Infof("[redis] Error searching geo. Err: %s ", err)
	}

	geos := make([]Geo, len(res))
	for idx, g := range res {
		geos[idx] = Geo{
			UUID:    g.Name,
			Lat:     g.Latitude,
			Lng:     g.Longitude,
			DistKm:  g.Dist,
			GeoHash: g.GeoHash,
		}
	}

	return geos
}

func (r RedisCtx) GetAllGeos() []Geo {
	return r.SearchLocation(r.SearchLocationQuery(0, 0, 21000))
}

// Remove all geos
func (r RedisCtx) ClearGeo() {
	ctx := context.Background()
	res, _ := r.rdb.Del(ctx, locations).Result()
	utils.PP(res)
}

// --------------------------------------------------------------------------------
// ClientGeo

func (r RedisCtx) GetClientGeo(uuid string) *ClientGeo {
	return &ClientGeo{
		Client: *r.GetClient(uuid),
		Geo:    *r.GetGeo(uuid),
	}
}

func (r RedisCtx) GetClientGeos(pattern string) []ClientGeo {
	ctx := context.Background()
	var clients []ClientGeo
	iter := r.rdb.Scan(ctx, 0, ClientPrefix+pattern, 0).Iterator()
	for iter.Next(ctx) {
		clients = append(clients, *r.GetClientGeo(iter.Val()))
	}
	return clients
}

func (r RedisCtx) DeleteClientGeo(uuid string) string {
	r.DelGeo(uuid)
	return r.DeleteClient(uuid)
}

func (r RedisCtx) DeleteClientGeos(uuids []string) []string {
	r.DelGeos(uuids)
	return r.DeleteClients(uuids)
}
