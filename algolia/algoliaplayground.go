package algolia

import (
	"fmt"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	// "github.com/michelemendel/beehlp-backend/algolia/indexbeehlp"

	u "github.com/michelemendel/bhbe/utils"
)

func mainAlgolia() {
	algoliaIndexName := IndexBeehlp{}
	index := InitIndex(algoliaIndexName)

	// someCreation(index)
	// someUpdate(index)
	// someDelete(index)
	// indexbeehlp.ClearIndex(index)
	someSearch(index)
}

func someCreation(index *search.Index) {
	locations := []Location{
		// indexbeehlp.MakeNewLocation(59.9138688, 10.7522451),
		// indexbeehlp.MakeNewLocation(59.9109398, 10.7387413),
		// indexbeehlp.MakeNewLocation(59.9106999, 10.7352777),
	}
	u.PP(locations)
	UpdateLocations(index, locations)
}

func someUpdate(index *search.Index) {
	locations := []Location{
		// indexbeehlp.MakeUpdateLocation("127214000", "2IxerbwIRxFOThAUZmVw7P5Im7B", "John", 59.10, 10.10),
		// indexbeehlp.MakeUpdateLocation("127143000", 59.9109398, 10.7387413),
	}
	u.PP(locations)
	UpdateLocations(index, locations)
}

func someDelete(index *search.Index) {
	objectIDs := []string{"127144111", "14999999999"}
	fmt.Println("Objects to delete", objectIDs)
	DeleteLocations(index, objectIDs)
}

func someSearch(index *search.Index) {
	// algoliaIndex1 := airports.IndexAirports{}
	// index1 := algolia.InitIndex(algoliaIndex1)
	// searchRes1 := latlngSearch(index1)
	// searchRes1 := ipSearch(index1)
	// fmt.Println("Nof hits", searchRes1.NbHits)
	// pp.PP(searchRes1)
	// pp.PP(algoliaIndex1.AirportsGetPlace(searchRes1))

	// searchRes2 := latlngSearch(index)
	// searchRes2 := ipSearch(index2)
	// fmt.Println("Nof hits", searchRes2.NbHits)
	// pp.PP(searchRes2)
	// u.PP(BeehlpGetPlaces(searchRes2))
}
