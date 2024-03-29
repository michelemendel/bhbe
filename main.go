package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/michelemendel/bhbe/redis"
	"github.com/michelemendel/bhbe/server"
	"github.com/michelemendel/bhbe/utils"
	"go.uber.org/zap"
)

var lg *zap.SugaredLogger

func init() {
	lg = utils.Log()
	serverEnvFile := utils.ResolveFilename("", ".env")
	err := godotenv.Load(serverEnvFile)
	if err != nil {
		lg.Panic("[main] Error loading file ", serverEnvFile)
	}
}

func main() {
	hostAddr := os.Getenv("EXTERNAL_ADDRESS")
	apiServerPort := os.Getenv("API_SERVER_PORT")
	server.StartApiServer(hostAddr, apiServerPort)
}

func main1() {
	redCtx := redis.InitRedisClient()

	// createSomeClientsWithGeos(redCtx)
	// uuid := "client:2LXmzMC6WkIyll2f04xMENSHyrK"

	// redCtx.DeleteClient(uuid)
	// redCtx.UpsertClient(uuid, "ArneBobMartin")

	// utils.PP(redCtx.GetClient(uuid))
	// utils.PP(redCtx.GetClientGeo(uuid))

	// --------------------------------------------

	// redCtx.ClearGeo()
	// utils.PP(redCtx.DelClientGeo(uuid))
	utils.PP(redCtx.SearchLocation(redCtx.SearchLocationQuery(12.511, 41.911, 30)))
	// utils.PP(redCtx.GetAllGeos())
	// redCtx.UpsertGeo(uuid, 12.511, 41.911)

	// --------------------------------------------

	// utils.PP(redCtx.DeleteClientGeos(redCtx.GetKeys("*")))
	// utils.PP(redCtx.GetClients("*"))
	// utils.PP(redCtx.GetClientGeos("*"))
	// redCtx.DeleteKeys(redCtx.GetKeys("*"))
	// utils.PP(redCtx.GetKeys("*"))

}

func createSomeClientsWithGeos(redCtx *redis.RedisCtx) {
	prefix := "client:"
	uuid := redCtx.UpsertClient(prefix+utils.GenerateUUID(), "Arne")
	redCtx.UpsertGeo(uuid, 12.511, 41.911)
	uuid = redCtx.UpsertClient(prefix+utils.GenerateUUID(), "Bob")
	redCtx.UpsertGeo(uuid, 12.522, 41.922)
	uuid = redCtx.UpsertClient(prefix+utils.GenerateUUID(), "Carl")
	redCtx.UpsertGeo(uuid, 12.533, 41.933)
}
