package main

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/michelemendel/bhbe/algolia"
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

func main1() {
	algoliaIndexName := algolia.IndexBeehlp{}
	algoliaIndex := algolia.InitIndex(algoliaIndexName)
	server.StartApiServer(hostAddr, apiServerPort, algoliaIndex)
}

func main() {
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
	// utils.PP(redCtx.SearchLocation(redCtx.SearchLocationQuery(12.511, 41.911, 30)))
	// utils.PP(redCtx.GetAllGeos())
	// redCtx.UpsertGeo(uuid, 12.511, 41.911)

	// --------------------------------------------

	// utils.PP(redCtx.DeleteClientGeos(redCtx.GetKeys("client:*")))
	// utils.PP(redCtx.GetClients("*"))
	utils.PP(redCtx.GetClientGeos("*"))
	// utils.PP(redCtx.GetKeys("client:*"))

}

func createSomeClientsWithGeos(redCtx *redis.RedisCtx) {
	uuid := redCtx.CreateClient("client:", "Arne")
	redCtx.UpsertGeo(uuid, 12.511, 41.911)
	uuid = redCtx.CreateClient("client:", "Bob")
	redCtx.UpsertGeo(uuid, 12.522, 41.922)
	uuid = redCtx.CreateClient("client:", "Carl")
	redCtx.UpsertGeo(uuid, 12.533, 41.933)
}
