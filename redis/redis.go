package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/michelemendel/bhbe/data"
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

func (r RedisCtx) DeleteKeys(keys []string) []string {
	ctx := context.Background()
	r.rdb.Del(ctx, keys...)
	return keys
}

func (r RedisCtx) DeleteALl() {
	ctx := context.Background()
	r.rdb.FlushDB(ctx)
}

// --------------------------------------------------------------------------------
// Client

func (r RedisCtx) setClient(uuid string, name string) string {
	ctx := context.Background()

	if name == "" {
		name = "Anonymous"
	}

	err := r.rdb.HSet(ctx, string(uuid), fUUID, uuid, fName, name, fUpdatedAt, utils.Now()).Err()
	if err != nil {
		lg.Warnf("[redis] Error setting client: %s. Err: %s", uuid, err)
	}
	return uuid
}

func (r RedisCtx) updateClientGeoUpdateAt(uuid string) string {
	ctx := context.Background()

	err := r.rdb.HSet(ctx, string(uuid), fGeoUpdatedAt, utils.Now()).Err()
	if err != nil {
		lg.Warnf("[redis] Error setting client: %s. Err: %s", uuid, err)
	}
	return uuid
}

func (r RedisCtx) UpsertClient(uuid string, name string) string {
	return r.setClient(uuid, name)
}

func (r RedisCtx) DeleteClient(uuid string) string {
	return r.DeleteKeys([]string{uuid})[0]
}

func (r RedisCtx) DeleteClients(uuids []string) []string {
	return r.DeleteKeys(uuids)
}

func (r RedisCtx) GetClient(uuid string) *data.Client {
	ctx := context.Background()

	// Scan can't scan time
	// err := r.rdb.HMGet(ctx, uuid, fName, fUpdatedAt).Scan(&client)
	client, err := r.rdb.HMGet(ctx, string(uuid), fUUID, fName, fUpdatedAt, fGeoUpdatedAt).Result()
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

	return &data.Client{
		UUID:         client[0].(string),
		Name:         client[1].(string),
		UpdatedAt:    updatedAt,
		GeoUpdatedAt: geoUpdatedAt,
	}
}

func (r RedisCtx) GetClients(pattern string) []data.Client {
	ctx := context.Background()
	var clients []data.Client
	iter := r.rdb.Scan(ctx, 0, ClientPrefix+pattern, 0).Iterator()
	for iter.Next(ctx) {
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

func (r RedisCtx) GetGeo(uuid string) *data.Geo {
	ctx := context.Background()
	res, err := r.rdb.GeoPos(ctx, locations, string(uuid)).Result()
	if err != nil {
		lg.Infof("[redis] Error getting geo for UUID %s. Err: %s ", uuid, err)
	}

	if res[0] == nil {
		return &data.Geo{
			UUID: uuid,
			Lat:  0,
			Lng:  0,
		}
	}

	return &data.Geo{
		UUID: uuid,
		Lat:  res[0].Latitude,
		Lng:  res[0].Longitude,
	}
}

func (r RedisCtx) DeleteGeo(uuid string) string {
	ctx := context.Background()
	_, err := r.rdb.ZRem(ctx, locations, uuid).Result()
	if err != nil {
		lg.Infof("[redis] Error deleting geo for UUID %s. Err: %s ", uuid, err)
	}
	return uuid
}

// Deletes multiple geos in key locations
func (r RedisCtx) DeleteGeos(uuids []string) []string {
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

func (r RedisCtx) SearchLocation(query *redis.GeoSearchLocationQuery) []*data.Geo {
	ctx := context.Background()
	res, err := r.rdb.GeoSearchLocation(ctx, locations, query).Result()
	if err != nil {
		lg.Infof("[redis] Error searching geo. Err: %s ", err)
	}

	geos := make([]*data.Geo, len(res))
	for idx, g := range res {
		geos[idx] = &data.Geo{
			UUID:    g.Name,
			Lat:     g.Latitude,
			Lng:     g.Longitude,
			DistKm:  g.Dist,
			GeoHash: g.GeoHash,
		}
	}

	return geos
}

func (r RedisCtx) GetAllGeos() []*data.Geo {
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

func (r RedisCtx) UpsertClientGeo(uuid, name string, lat, lng float64) string {
	r.UpsertClient(uuid, name)
	r.UpsertGeo(uuid, lat, lng)
	return uuid
}

func (r RedisCtx) GetClientGeo(uuid string) *data.ClientGeo {
	return &data.ClientGeo{
		Client: *r.GetClient(uuid),
		Geo:    *r.GetGeo(uuid),
	}
}

func (r RedisCtx) GetClientGeos(pattern string) []data.ClientGeo {
	ctx := context.Background()
	var clientGeos []data.ClientGeo
	iter := r.rdb.Scan(ctx, 0, ClientPrefix+pattern, 0).Iterator()
	for iter.Next(ctx) {
		clientGeos = append(clientGeos, *r.GetClientGeo(iter.Val()))
	}
	return clientGeos
}

func (r RedisCtx) DeleteClientGeo(uuid string) string {
	r.DeleteGeo(uuid)
	return r.DeleteClient(uuid)
}

func (r RedisCtx) DeleteClientGeos(uuids []string) []string {
	r.DeleteGeos(uuids)
	return r.DeleteClients(uuids)
}
