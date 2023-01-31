package main

import (
	"flag"
	"net"

	"github.com/joho/godotenv"
	"github.com/michelemendel/bhbe/algolia"
	"github.com/michelemendel/bhbe/server"
	"github.com/michelemendel/bhbe/utils"
	"go.uber.org/zap"
)

var lg *zap.SugaredLogger
var hostAddr, apiServerPort string
var isProductionMode bool

func init() {
	lg = utils.Log()
	serverEnvFile := utils.ResolveFilename("", ".env")
	err := godotenv.Load(serverEnvFile)
	if err != nil {
		lg.Fatal("[main] Error loading file ", serverEnvFile)
	}

	// usage (don't use port 6666)
	// > run --host 212.251.202.26
	flag.StringVar(&hostAddr, "host", GetIP(), "Host address")
	flag.StringVar(&apiServerPort, "apiport", "8588", "Port for the API server")
	flag.BoolVar(&isProductionMode, "production", false, "Production mode")
	flag.Parse()
}

func main() {
	algoliaIndexName := algolia.IndexBeehlp{}
	algoliaIndex := algolia.InitIndex(algoliaIndexName)
	server.StartApiServer(hostAddr, apiServerPort, algoliaIndex)
}

func GetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		lg.Error("Couldn't get an IP address", err)
		return ""
	}
	for _, address := range addrs {
		// Check the address type and if it is not a loopback then display it
		ipnet, ok := address.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	lg.Error("Couldn't get an IP address", err)
	return ""
}
