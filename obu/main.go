package main

// On Board Unit (OBU)

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
)

type onBoardUnit struct {
	Id       string  `json:"id"`
	Spz      string  `json:"spz"`
	Country  string  `json:"country"`
	Credit   float64 `json:"credit"`
	Currency string  `json:"currency"`
	Weight   int     `json:"weight"`
	Emission string  `json:"emission"`
	Category string  `json:"category"`
	Axles    int     `json:"axles"`
}

type wptRecords struct {
	LatRad    []float64 `json:"latRad"`
	LonRad    []float64 `json:"lonRad"`
	Distances []float64 `json:"distances"`
	Version   string    `json:"version"`
	Len       int       `json:"len"`
	Name      string    `json:"name"`
	Checksum  string    `json:"checksum"`
}

type polygon struct {
	I    []int    `json:"i"` // i - road
	J    []int    `json:"j"` // j - point of the road
	Time []string `json:"time"`
}

type ticket struct {
	Obu         onBoardUnit `json:"obu"`
	CheckPoints polygon     `json:"polygon"`
}

const CACHE_DIR = "cache"
const URL_SERVER = "http://localhost:8905"
const EARTH_RADIUS = 6371000 // Radius of the Earth in meters
const THRESHOLD = 20         // Threshold for algorithm in meters
const OBU_NAME = "obu1"

var model []wptRecords
var route wptRecords
var obu onBoardUnit

func main() {
	obuName := flag.String("name", OBU_NAME, "OBU name")
	flag.Parse()

	err := readJson(fmt.Sprintf("%s.json", *obuName), &obu)
	if err != nil {
		return
	}
	err = initObu(URL_SERVER, &obu)
	if err != nil {
		fmt.Printf("Cannot initialized OBU with the server %s\n%v", URL_SERVER, err)
		return
	}

	err = readGpx(fmt.Sprintf("%s.gpx", *obuName), &route)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	getGeoModel(URL_SERVER, &model)
	if len(model) == 0 {
		fmt.Printf("Model is not loaded either from cache nor %s", URL_SERVER)
		return
	}

	var checkPoints polygon
	driveAlgorithm(route, &checkPoints)
	if len(checkPoints.Time) == 0 {
		fmt.Println("No toll road detected")
		return
	}
	sendTicket(URL_SERVER, checkPoints, obu)
	// fmt.Println(model[0].LatRad)
}

func driveAlgorithm(route wptRecords, checkPoints *polygon) {
	//	var distance float64
	var nearest_i int
	var nearest_j int

	for i := 0; i < route.Len; i++ {
		t := time.Now()
		calcDistanceToModel(route.LatRad[i], route.LonRad[i])

		shortestDistanceInModel(&nearest_i, &nearest_j)

		if findPair(checkPoints.I, checkPoints.J, nearest_i, nearest_j) == -1 && model[nearest_i].Distances[nearest_j] <= THRESHOLD { // Check if point is not already in array
			// Check if distance is within THRESHOLD to be evaluated as paid road
			checkPoints.I = append(checkPoints.I, nearest_i) // Road section
			checkPoints.J = append(checkPoints.J, nearest_j) // Point of the road
			checkPoints.Time = append(checkPoints.Time, t.Format(time.RFC3339))
		}
		//		fmt.Println(nearest_i, nearest_j)
		//		fmt.Println(model[nearest_i].Distances[nearest_j])
	}
	fmt.Println(*checkPoints)
}

// Calculate distance from current latRad and lonRad, return point in Threshold and shortest distance
func calcDistanceToModel(latRad float64, lonRad float64) {
	for i, v := range model {
		for j := 0; j < v.Len; j++ {
			distance := haversine(latRad, lonRad, v.LatRad[j], v.LonRad[j])
			model[i].Distances[j] = distance
		}
	}
}

func shortestDistanceInModel(index_i, index_j *int) {
	result := model[0].Distances[0]
	for i, v := range model {
		for j := 0; j < v.Len; j++ {
			if result > v.Distances[j] {
				result = v.Distances[j]
				*index_i = i
				*index_j = j
			}
		}

	}
}

func readJson(filename string, v any) error {
	result, err := os.Open(filename)
	if os.IsNotExist(err) {
		fmt.Printf("error: %s does not exist.", filename)
		return err
	}
	if err != nil {
		fmt.Printf("error %s", err)
		return err
	}
	defer result.Close()
	byteValue, _ := ioutil.ReadAll(result)
	json.Unmarshal(byteValue, &v)

	return nil
}

func readGpx(filename string, route *wptRecords) error {
	result, err := os.Open(filename)
	if os.IsNotExist(err) {
		fmt.Printf("error: File %s does not exist.", filename)
		return err
	}
	if err != nil {
		fmt.Printf("error %s", err)
		return err
	}
	defer result.Close()

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(filename); err != nil {
		fmt.Println(err)
		return err
	}
	root := doc.SelectElement("gpx")
	for _, e := range root.SelectElements("wpt") {
		latStr := e.SelectAttrValue("lat", "0.0")
		latDeg, _ := strconv.ParseFloat(latStr, 64)
		latRad := degreesToRadians(latDeg)
		route.LatRad = append(route.LatRad, latRad)

		lonStr := e.SelectAttrValue("lon", "0.0")
		lonDeg, _ := strconv.ParseFloat(lonStr, 64)
		lonRad := degreesToRadians(lonDeg)
		route.LonRad = append(route.LonRad, lonRad)
	}
	if len(route.LatRad) == len(route.LonRad) {
		route.Len = len(route.LatRad)
	} else {
		return fmt.Errorf("Error: Inconsistency length of arrays LatRad and LonRad")
	}
	return err
}

func getGeoModel(urlServer string, model *[]wptRecords) {
	url := fmt.Sprintf("%s/geomodel", urlServer)
	var filename string = "model.json"
	if uptodate(filepath.Join(CACHE_DIR, filename), url, model) {
		return
	}

	result := newRequest(url)
	if len(result) == 0 {
		fmt.Printf("Error: output is empty")
		return
	}
	byteResult := []byte(result)
	if json.Valid(byteResult) {
		json.Unmarshal(byteResult, model)
		writeFile(CACHE_DIR, filename, result)
	} else {
		fmt.Printf("Error: Model from %s is not valid json", url)
	}
}

func uptodate(filename, url string, model *[]wptRecords) bool {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("error %s", err)
		return false
	}
	json.Unmarshal(data, model)
	var versions []wptRecords
	urlVersion := fmt.Sprintf("%s?v", url)
	result := newRequest(urlVersion)
	if result == "" {
		return false
	}
	json.Unmarshal([]byte(result), &versions)

	for i := 0; i < len(versions); i++ {
		d := fmt.Sprintf("%+v", (*model)[i])
		h := hash([]byte(d))

		if versions[i].Version != (*model)[i].Version ||
			versions[i].Checksum != h {
			return false
		}
	}
	fmt.Println("Geomodel is correct and up to date.")
	return true

}

func initObu(urlServer string, obu *onBoardUnit) error {
	byteResult, _ := json.Marshal(obu)
	payload := strings.NewReader(string(byteResult))
	url := fmt.Sprintf("%s/obu", urlServer)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Set("Content-Type", "application/json")
	content, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	value, _ := io.ReadAll(content.Body)
	err = json.Unmarshal(value, obu)
	if err != nil {
		return err
	}
	return nil
}

func sendTicket(urlServer string, checkPoints polygon, obu onBoardUnit) {
	url := fmt.Sprintf("%s/ticket", urlServer)
	var t ticket
	t.CheckPoints = checkPoints
	t.Obu = obu

	byteResult, _ := json.Marshal(t)
	// fmt.Printf("%s", string(byteResult))
	payload := strings.NewReader(string(byteResult))
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Set("Content-Type", "application/json")
	content, err := http.DefaultClient.Do(req)
	if err == nil {
		fmt.Println("Ticket sent.")
		value, _ := io.ReadAll(content.Body)
		fmt.Println(string(value))
	}

}

func newRequest(url string) string {
	var answer string = ""
	//	t := time.Now().Add(2 * time.Second)
	//	ctx, cancel := context.WithCancel(context.TODO())
	client := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 2 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   2 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	reqm, _ := http.NewRequest("GET", url, nil)
	reqm.Header.Set("User-Agent", "Mozilla")
	reqm.Header.Set("Content-Type", "text/html")
	content, err := client.Do(reqm)
	if err != nil {
		fmt.Println(err)
		if content != nil {
			fmt.Println("statusCode: ", content.StatusCode)
		}
		return answer
	} else if content.StatusCode >= 400 {
		return answer
	}

	value, err := io.ReadAll(content.Body)
	if err != nil {
		fmt.Println(err)
		return answer
	}
	answer = string(value)
	return answer
}

func writeFile(dir, filename, value string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
		if err != nil {
			fmt.Printf("error %s", err)
		}
	}
	err := os.WriteFile(dir+"/"+filename, []byte(value), 0644)
	if err != nil {
		fmt.Printf("error %s", err)
	}
}

func findPair(array1, array2 []int, value1, value2 int) int {
	var index int = -1
	if len(array1) != len(array2) {
		fmt.Printf("error: Array1 and Array2 dont have the same length.")
		return index
	}
	var i = findValue(array1, value1)
	if i == findValue(array2, value2) {
		index = i
	}
	return index
}

func findValue(array []int, value1 int) int {
	var index int = -1
	for i := 0; i < len(array); i++ {
		if array[i] == value1 {
			index = i
		}
	}
	return index
}

// Haversine calculates the distance between two GPS coordinates using the Haversine formula.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {

	dlat := lat2 - lat1
	dlon := lon2 - lon1

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := EARTH_RADIUS * c

	return distance
}

func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func hash(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
