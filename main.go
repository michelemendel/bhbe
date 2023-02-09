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

	// uuid := redCtx.CreateClient("John")

	// uuid := "2LV2cDphbehdMIBFjNdv9CcXXrg"
	// fmt.Printf("UUID: %s\n", uuid)

	// uuid = redCtx.DeleteClient(uuid)
	// uuid = redCtx.UpdateClient(uuid, "Arne")

	// client := redCtx.GetClient(uuid)
	// utils.PP(client)

	// --------------------------------------------

	// redCtx.ClearGeo()
	// redCtx.AddGeo("A1", 12.5, 41.9)
	// redCtx.AddGeo("A2", 12.55, 41.99)

	// utils.PP(redCtx.DelGeo("A1"))
	// utils.PP(redCtx.GetGeo("A1"))

	slq := redCtx.SearchLocationQuery(12.555, 41.999, 2)
	utils.PP(redCtx.SearchLocation(slq))
	// utils.PP(redCtx.GetAllGeos())

}
