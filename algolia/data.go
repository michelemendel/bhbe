package algolia

import u "github.com/michelemendel/bhbe/utils"

// --------------------------------------------------------------------------------
// Beehlp: Structs for the Algolia index search result

type SearchResult struct {
	AroundLatLng string `json:"latlng"` //Only used when using an IP address. It returns the coordinates.
	Locations    []Location
}

type Location struct {
	ErrorMsg   string      `json:"errorMsg"`
	ObjectID   string      `json:"objectID"`
	ClientUuid string      `json:"clientUuid"`
	Name       string      `json:"name"`
	UpdatedAt  u.Timestamp `json:"updatedAt"`
	GeoLoc     GeoLoc      `json:"_geoloc"`
}

type GeoLoc struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// ----------------------------------------
// Beehlp: Helper functions

func MakeNewLocation(clientUuid, name string) Location {
	ret := Location{
		ClientUuid: clientUuid,
		Name:       name,
		UpdatedAt:  u.StampTimeNow(),
	}
	return ret
}

func MakeUpdateLocation(objectId, clientUuid, username string, lat, lng float64) Location {
	return Location{
		ObjectID:   objectId,
		ClientUuid: clientUuid,
		Name:       username,
		UpdatedAt:  u.StampTimeNow(),
		GeoLoc:     GeoLoc{Lat: lat, Lng: lng},
	}
}

func MakeErrorLocation(errorMsg string) Location {
	return Location{
		ErrorMsg: errorMsg,
	}
}

// ----------------------------------------
// Beehlp: Structs to store data in the Algolia index
type Client struct {
	ClientUuid string  `json:"clientUuid"`
	ObjectID   string  `json:"objectID"` // Algolia's unique identifier
	Name       string  `json:"name"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
}

type DeleteUser struct {
	ClientUuid string `json:"clientUuid"`
}
