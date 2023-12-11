package server

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/beevik/etree"
)

const earthRadius = 6371000 // Radius of the Earth in meters

type WptRecords struct {
	LatRad    []float64 `json:"latRad"`
	LonRad    []float64 `json:"lonRad"`
	Distances []float64 `json:"distances"`
	Version   string    `json:"version"`
	Len       int       `json:"len"`
	Name      string    `json:"name"`
	Checksum  string    `json:"checksum"`
}

type Polygon struct {
	I    []int    `json:"i"` // i - road
	J    []int    `json:"j"` // j - point of the road
	Time []string `json:"time"`
}

var Model []WptRecords

func LoadModel() {
	var model []WptRecords
	var route1 WptRecords
	readGpx("model/i35.gpx", &route1)
	model = append(model, route1)

	var route2 WptRecords
	readGpx("model/d10.gpx", &route2)
	model = append(model, route2)
	Model = model
}

func readGpx(filename string, route *WptRecords) {
	result, err := os.Open(filename)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		fmt.Printf("error %s", err)
		return
	}
	defer result.Close()

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(filename); err != nil {
		fmt.Println(err)
		return
	}
	root := doc.SelectElement("gpx")
	name := root.SelectElement("title").Text()
	route.Name = name
	route.Version = root.SelectElement("version").Text()

	for _, e := range root.SelectElements("wpt") {
		latStr := e.SelectAttrValue("lat", "0.0")
		latDeg, _ := strconv.ParseFloat(latStr, 64)
		latRad := degreesToRadians(latDeg)
		route.LatRad = append(route.LatRad, latRad)

		lonStr := e.SelectAttrValue("lon", "0.0")
		lonDeg, _ := strconv.ParseFloat(lonStr, 64)
		lonRad := degreesToRadians(lonDeg)
		route.LonRad = append(route.LonRad, lonRad)
		route.Distances = append(route.Distances, 0.0)
	}
	route.Len = len(route.LonRad)

}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// Haversine calculates the distance between two GPS coordinates using the Haversine formula.
// parameters are in degrees
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {

	dlat := lat2 - lat1
	dlon := lon2 - lon1

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := earthRadius * c

	return distance
}
