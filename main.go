package main

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/michelemendel/bhbe/redis"
	"github.com/michelemendel/bhbe/server"
	"github.com/michelemendel/bhbe/utils"
	"go.uber.org/zap"
)

var lg *zap.SugaredLogger
var hostAddr, apiServerPort string

func init() {
	lg = utils.Log()
	serverEnvFile := utils.ResolveFilename("", ".env")
	err := godotenv.Load(serverEnvFile)
	if err != nil {
		lg.Fatal("[main] Error loading file ", serverEnvFile)
	}

	// > run --host 212.251.202.26
	flag.StringVar(&hostAddr, "host", os.Getenv("EXTERNAL_ADDRESS"), "Host address")
	flag.StringVar(&apiServerPort, "apiport", "8588", "Port for the API server")
	flag.Parse()
}

func main() {
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
