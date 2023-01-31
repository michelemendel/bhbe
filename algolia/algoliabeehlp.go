package algolia

import (
	"os"
	"strconv"
	"sync"

	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	u "github.com/michelemendel/bhbe/utils"
)

// OpenAI
// Prompt:
// Golang get 3 geolocations in Sao Paolo based on the following struct:
// type Geolocation struct {
// 	Lat  float64 `json:"lat"`
// 	Lng  float64 `json:"lng"`
// }

type IndexBeehlp struct{}

type recipient struct {
	ClientUuid string  `json:"clientUuid"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
}

var mutex sync.Mutex

// Returns an Algolia objectID
func Create(index *search.Index, clientUuid, name string) string {
	objectId := GetObjectIdByClientUuid(index, clientUuid)
	if objectId != "" {
		lg.Warnf("[handlers algolia] Object already exists: %s", objectId)
		return objectId
	}
	locations := []Location{
		MakeNewLocation(clientUuid, name),
	}
	return UpdateLocations(index, locations)[0]
}

// Returns an Algolia objectID
func Update(index *search.Index, objectID, clientUuid, username string, lat, lng float64) string {
	mutex.Lock()
	defer mutex.Unlock()

	if objectID == "" {
		objectID = GetObjectIdByClientUuid(index, clientUuid)
	}
	if objectID == "" {
		lg.Warnf("[handlers algolia] Object with id %s doesn't exist. Creating a new one.", objectID)
		return Create(index, clientUuid, username)
	}
	locations := []Location{
		MakeUpdateLocation(objectID, clientUuid, username, lat, lng),
	}
	return UpdateLocations(index, locations)[0]
}

func DeleteLocation(index *search.Index, objectID, clientUUID string) {
	if objectID == "" {
		lg.Warnf("[handlers algolia] Object with id %s doesn't exist. Loking in db.", objectID)
		objectID = GetObjectIdByClientUuid(index, clientUUID)
	}
	objectIDs := []string{objectID}
	lg.Infof("[handlers algolia] Objects to delete: %s", objectIDs)
	DeleteLocations(index, objectIDs)
}

func Recipient(index *search.Index, toUuid string) *recipient {
	searchRes := SearchByQueryExact(index, toUuid)
	if searchRes.NbHits == 0 {
		return &recipient{}
	}
	recipient := GetRecipients(searchRes)[0]
	return recipient
}

// Get recipients from Algolia, either by lat/lng or by IP
// radius in km
func Recipients(index *search.Index, lat, lng float64, radius int) []*recipient {
	searchRes := LatlngSearch(index, lat, lng, radius)
	recipients := GetRecipients(searchRes)
	return recipients
}

// Get objectId by clientUuid
func GetObjectIdByClientUuid(index *search.Index, clientUuid string) string {
	searchRes := SearchByQueryExact(index, clientUuid)
	if searchRes.NbHits == 0 {
		return ""
	}
	return searchRes.Hits[0]["objectID"].(string)
}

// --------------------------------------------------------------------------------
// Algolia interface functions
func (IndexBeehlp) SearchableAttributes() []string {
	return []string{"_geoloc", "clientUuid"}
}

func (IndexBeehlp) CustomRanking() []string {
	return []string{""}
}

func (IndexBeehlp) DbIndex() string {
	return os.Getenv("ALGOLIA_INDEX_BEEHLP")
}

// --------------------------------------------------------------------------------
// Create/Update
func UpdateLocations(index *search.Index, locations []Location) []string {
	res, err := index.SaveObjects(locations)
	if err != nil {
		lg.Errorf("[%s] Couldn't update object", index.GetName())
	}
	return res.ObjectIDs()
}

// --------------------------------------------------------------------------------
// Delete
func DeleteLocations(index *search.Index, objectIDs []string) {
	_, err := index.DeleteObjects(objectIDs)
	if err != nil {
		lg.Warnf("[%s] Couldn't delete locations %v", index.GetName(), objectIDs)
	}
}

// --------------------------------------------------------------------------------
// Clear index
func ClearIndex(index *search.Index) {
	_, err := index.ClearObjects()
	if err != nil {
		lg.Errorf("[%s]Couldn't clear index", index.GetName())
	}
	lg.Infof("[%s] Index is cleared.", index.GetName())
}

// --------------------------------------------------------------------------------
// Search

// Radius in km
func LatlngSearch(index *search.Index, lat, lng float64, radius int) search.QueryRes {
	query := ""
	searchRes := SearchByLatLngRadius(index, query, strconv.FormatFloat(lat, 'f', 5, 64), strconv.FormatFloat(lng, 'f', 5, 64), radius*1000)
	return searchRes
}

// Radius in km
func IpSearch(index *search.Index, ip string, radius int) search.QueryRes {
	query := ""
	searchRes := SearchByIpRadius(index, query, ip, radius*1000)
	return searchRes
}

func GetRecipients(algoliaSearchRes search.QueryRes) []*recipient {
	places := GetLocations(algoliaSearchRes)
	recipients := []*recipient{}
	for _, loc := range places.Locations {
		recipients = append(recipients, &recipient{
			ClientUuid: loc.ClientUuid,
			Lat:        loc.GeoLoc.Lat,
			Lng:        loc.GeoLoc.Lng,
		})
	}

	return recipients
}

func GetLocations(searchRes search.QueryRes) SearchResult {
	var places []Location

	for _, h := range searchRes.Hits {
		geoLoc, ok := h["_geoloc"].(map[string]interface{})
		if !ok {
			msg := "Problems asserting geoloc result from searchRes.Hits"
			lg.Info(msg)
			places = append(places, MakeErrorLocation(msg))
			continue
		}

		lat, ok := geoLoc["lat"].(float64)
		if !ok {
			msg := "Problems asserting lat result from searchRes.Hits"
			lg.Info(msg)
			places = append(places, MakeErrorLocation(msg))
			continue
		}

		lng, ok := geoLoc["lng"].(float64)
		if !ok {
			msg := "Problems asserting lng result from searchRes.Hits"
			lg.Info(msg)
			places = append(places, MakeErrorLocation(msg))
			continue
		}

		objectId, ok := h["objectID"].(string)
		if !ok {
			msg := "Problems asserting objectId result from searchRes.Hits"
			lg.Info(msg)
			places = append(places, MakeErrorLocation(msg))
			continue
		}

		clientUuid, ok := h["clientUuid"].(string)
		if !ok {
			msg := "Problems asserting clientUuid result from searchRes.Hits"
			lg.Info(msg)
			places = append(places, MakeErrorLocation(msg))
			continue
		}

		p := Location{
			ObjectID:   objectId,
			ClientUuid: clientUuid,
			UpdatedAt:  u.StampTimeNow(),
			GeoLoc: GeoLoc{
				Lat: lat,
				Lng: lng,
			},
		}

		places = append(places, p)
	}

	return SearchResult{
		AroundLatLng: searchRes.AroundLatLng,
		Locations:    places,
	}
}
