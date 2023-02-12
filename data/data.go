package data

import "time"

// --------------------------------------------------------------------------------
// Server
type Connections struct {
	UUID       string `json:"uuid"`
	RemoteAddr string `json:"remoteaddr"`
}

// --------------------------------------------------------------------------------
// Redis
type Client struct {
	UUID         string    `json:"uuid"`
	Name         string    `json:"name"`
	UpdatedAt    time.Time `json:"updated"`
	GeoUpdatedAt time.Time `json:"geoupdatedat"`
}

type Geo struct {
	UUID    string  `json:"uuid"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	DistKm  float64 `json:"distkm"`
	GeoHash int64   `json:"geohash"`
}

type ClientGeo struct {
	Client Client `json:"client"`
	Geo    Geo    `json:"geo"`
}

// --------------------------------------------------------------------------------
type ConnsRedisInfo struct {
	Connections []Connections `json:"connections"`
	ClientGeo   []ClientGeo   `json:"clientgeo"`
}
