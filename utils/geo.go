package utils

import (
	"math"
)

// :::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
// :::                                                                         :::
// :::  This routine calculates the distance between two points (given the     :::
// :::  latitude/longitude of those points). It is based on free code used to  :::
// :::  calculate the distance between two locations using GeoDataSource(TM)   :::
// :::  products.                                                              :::
// :::                                                                         :::
// :::  Definitions:                                                           :::
// :::    South latitudes are negative, east longitudes are positive           :::
// :::                                                                         :::
// :::  Passed to function:                                                    :::
// :::    lat1, lon1 = Latitude and Longitude of point 1 (in decimal degrees)  :::
// :::    lat2, lon2 = Latitude and Longitude of point 2 (in decimal degrees)  :::
// :::    optional: unit = the unit you desire for results                     :::
// :::           where: 'M' is statute miles (default, or omitted)             :::
// :::                  'K' is kilometers                                      :::
// :::                  'N' is nautical miles                                  :::
// :::                                                                         :::
// :::  Worldwide cities and other features databases with latitude longitude  :::
// :::  are available at https://www.geodatasource.com                         :::
// :::                                                                         :::
// :::  For enquiries, please contact sales@geodatasource.com                  :::
// :::                                                                         :::
// :::  Official Web site: https://www.geodatasource.com                       :::
// :::                                                                         :::
// :::          Golang code James Robert Perih (c) All Rights Reserved 2018    :::
// :::                                                                         :::
// :::           GeoDataSource.com (C) All Rights Reserved 2017                :::
// :::                                                                         :::
// :::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::

// Some minor modifications by Michele Mendel

// Usage
//   winnipeg := coordinate{49.895077, -97.138451}
//   regina := coordinate{50.445210, -104.618896}
//   fmt.Println(distance(winnipeg.lat, winnipeg.lng, regina.lat, regina.lng))

type Coords struct {
	Lat float64
	Lng float64
}

// Distance in km
func Distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64) float64 {
	radlat1 := float64(math.Pi * lat1 / 180)
	radlat2 := float64(math.Pi * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(math.Pi * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)
	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / math.Pi
	dist = dist * 60 * 1.1515
	dist = dist * 1.609344 //km

	return dist
}

func MakeCoords(lat float64, lng float64) Coords {
	return Coords{lat, lng}
}
