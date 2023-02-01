package main

import (
	"flag"

	"github.com/joho/godotenv"
	"github.com/michelemendel/bhbe/algolia"
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
	flag.StringVar(&hostAddr, "host", utils.GetIP(), "Host address")
	flag.StringVar(&apiServerPort, "apiport", "8588", "Port for the API server")
	flag.Parse()
}

func main() {
	algoliaIndexName := algolia.IndexBeehlp{}
	algoliaIndex := algolia.InitIndex(algoliaIndexName)
	server.StartApiServer(hostAddr, apiServerPort, algoliaIndex)
}
