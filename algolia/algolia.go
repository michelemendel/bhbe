package algolia

import (
	"fmt"
	"os"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/michelemendel/bhbe/utils"
	"go.uber.org/zap"
)

// Algolia
// go get github.com/algolia/algoliasearch-client-go/v3
// https://www.algolia.com/doc/guides/getting-started/quick-start/tutorials/quick-start-with-the-api-client/go/?client=go
// https://github.com/algolia-samples/api-clients-quickstarts/tree/master/go
// https://www.algolia.com/doc/guides/managing-results/refine-results/geolocation/
// https://www.algolia.com/doc/guides/managing-results/refine-results/geolocation/how-to/filter-results-around-a-location/

var lg *zap.SugaredLogger

func init() {
	lg = utils.Log()
}

type Algolia interface {
	DbIndex() string
	SearchableAttributes() []string
	CustomRanking() []string
}

func InitIndex(algoliaIndex Algolia) *search.Index {
	// lg.Infof("[%s] Initialize Algolia", algoliaIndex.DbIndex())
	client := search.NewClient(os.Getenv("ALGOLIA_APP_ID"), os.Getenv("ALGOLIA_API_KEY"))
	index := client.InitIndex(algoliaIndex.DbIndex())

	_, err := index.SetSettings(search.Settings{
		SearchableAttributes: opt.SearchableAttributes(algoliaIndex.SearchableAttributes()...),
		AdvancedSyntax:       opt.AdvancedSyntax(true),
		// CustomRanking:        opt.CustomRanking(algoliaIndex.CustomRanking()...),
	})
	if err != nil {
		lg.Fatalf("Couldn't initiate Algolia search. Index %s", algoliaIndex.DbIndex())
	}

	return index
}

func SearchByLatLngRadius(index *search.Index, query string, lat, lng string, radius int) search.QueryRes {
	pos := lat + "," + lng
	// lg.Infof("SearchByLatLngRadius: %v,%v,%vkm", lat, lng, radius/1000)

	searchRes, err := index.Search(
		query,
		opt.AroundLatLng(pos),
		opt.AroundRadius(radius),
		opt.AroundPrecision(opt.AroundPrecisionRange{From: 0, Value: 200}),
	)
	if err != nil {
		lg.Errorf("Problems searching around a geolocation lat,lng: %s", pos)
	}

	return searchRes
}

func extraHeaders(ip string) map[string]string {
	return map[string]string{"X-Forwarded-For": ip}
}

func SearchByIpRadius(index *search.Index, query string, ip string, radius int) search.QueryRes {
	// lg.Info("Searching by IP", ip)

	searchRes, err := index.Search(
		query,
		opt.AroundLatLngViaIP(true),
		opt.ExtraHeaders(extraHeaders(ip)),
		opt.AroundRadius(radius),
		opt.AroundPrecision(opt.AroundPrecisionRange{From: 0, Value: 200}),
	)
	if err != nil {
		lg.Errorf("Problems searching by IP: %s", ip)
	}

	return searchRes
}

func SearchByQueryExact(index *search.Index, query string) search.QueryRes {
	// lg.Infof("SearchByQueryExact: %s", query)

	searchRes, err := index.Search(
		fmt.Sprintf("\"%s\"", query), //Exact match
		opt.AdvancedSyntax(true),
		opt.AdvancedSyntaxFeatures("exactPhrase"),
	)
	if err != nil {
		lg.Errorf("Problems searching by query: %s", query)
	}

	return searchRes
}
