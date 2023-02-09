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
	fName      = "name"
	fUpdatedAt = "updatedAt"
)

type Client struct {
	Name      string    `redis:"name"`
	UpdatedAt time.Time `redis:"time"`
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

func (r RedisCtx) setClient(uuid, name string) string {
	ctx := context.Background()

	if name == "" {
		name = "Anonymous"
	}

	err := r.rdb.HSet(ctx, uuid, fName, name, fUpdatedAt, utils.Now()).Err()
	if err != nil {
		lg.Warnf("[redis] Error setting client: %s", name)
		fmt.Println(err)
	}
	return uuid
}

// TODO: mendel Fix: Name should be unique
func (r RedisCtx) CreateClient(name string) string {
	uuid := utils.GenerateUUID()
	return r.setClient(uuid, name)
}

func (r RedisCtx) UpdateClient(uuid, name string) string {
	return r.setClient(uuid, name)
}

func (r RedisCtx) DeleteClient(uuid string) string {
	ctx := context.Background()
	r.rdb.Del(ctx, uuid)
	return uuid
}

func (r RedisCtx) GetClient(uuid string) *Client {
	ctx := context.Background()

	// var client Client

	// Scan can't scan time
	// err := r.rdb.HMGet(ctx, uuid, fName, fUpdatedAt).Scan(&client)
	client, err := r.rdb.HMGet(ctx, uuid, fName, fUpdatedAt).Result()
	if err != nil {
		lg.Infof("[redis] Error getting UUID %s. Err: %s ", uuid, err)
	}

	if client[0] == nil {
		client[0] = ""
	}

	if client[1] == nil {
		client[1] = ""
	}

	t, _ := time.Parse(time.RFC3339, client[1].(string))

	return &Client{
		Name:      client[0].(string),
		UpdatedAt: t,
	}
}

func (r RedisCtx) GetAllClients() {
	ctx := context.Background()
	var vals []string
	iter := r.rdb.Scan(ctx, 0, "", 0).Iterator()
	for iter.Next(ctx) {
		vals = append(vals, iter.Val())
	}

	utils.PP(vals)
}

// --------------------------------------------------------------------------------
// Geo

const locations = "locations"

type Geo struct {
	UUID    string
	Lat     float64
	Lng     float64
	Dist    float64
	GeoHash int64
}

func (r RedisCtx) AddGeo(UUID string, lat, lng float64) {
	ctx := context.Background()

	geoLocation := &redis.GeoLocation{
		Name:      UUID,
		Longitude: lng,
		Latitude:  lat,
	}

	res, _ := r.rdb.GeoAdd(ctx, locations, geoLocation).Result()
	fmt.Println("GeoAdd: ", res)
}

func (r RedisCtx) GetGeo(UUID string) Geo {
	ctx := context.Background()
	res, err := r.rdb.GeoPos(ctx, locations, UUID).Result()
	if err != nil {
		lg.Infof("[redis] Error getting geo for UUID %s. Err: %s ", UUID, err)
	}

	if res[0] == nil {
		return Geo{
			UUID: UUID,
			Lat:  0,
			Lng:  0,
		}
	}

	return Geo{
		UUID: UUID,
		Lat:  res[0].Latitude,
		Lng:  res[0].Longitude,
	}
}

func (r RedisCtx) DelGeo(UUID string) string {
	ctx := context.Background()
	res, err := r.rdb.ZRem(ctx, locations, UUID).Result()
	if err != nil {
		lg.Infof("[redis] Error deleting geo for UUID %s. Err: %s ", UUID, err)
	}

	//TODO: mendel remove
	fmt.Println(res)

	return UUID
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

	geos := make([]Geo, 0)
	for _, g := range res {
		geos = append(geos, Geo{
			UUID:    g.Name,
			Lat:     g.Latitude,
			Lng:     g.Longitude,
			Dist:    g.Dist,
			GeoHash: g.GeoHash,
		})
	}

	return geos
}

func (r RedisCtx) GetAllGeos() []Geo {
	return r.SearchLocation(r.SearchLocationQuery(0, 0, 21000))
}

func (r RedisCtx) ClearGeo() {
	ctx := context.Background()
	res, _ := r.rdb.Del(ctx, locations).Result()
	utils.PP(res)
}
