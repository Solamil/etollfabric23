package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Solamil/bp23/server"
)

type geoUrlParams struct {
	Version []string `json:"v"`
}

type ticket struct {
	Obu         server.OnBoardUnit `json:"obu"`
	CheckPoints server.Polygon     `json:"polygon"`
}

const PORT = 8905

var dbType string = "Blockchain"

func main() {
	server.LoadSazba()
	server.InitDb(dbType)
	port := flag.Int("port", PORT, "Port for the server to listen on.")

	http.HandleFunc("/", index_handler)
	http.HandleFunc("/obu", obu_handler)
	http.HandleFunc("/ticket", ticket_handler)
	http.HandleFunc("/geomodel", geo_handler)
	http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`The server for electronic toll road

	Usage:
	/ - This help
	/geomodel - Return geograhic model of toll roads.
	/geomodel?v - Return version and checksum of geographic model
	/ticket - Process driven toll roads given by OBUs and compute the toll.
	/obu - Initialize OBU and check information about OBU.

	Author michal.kukla@tul.cz
	2023
	`))

}

func ticket_handler(w http.ResponseWriter, r *http.Request) {
	t := &ticket{}
	if err := json.NewDecoder(r.Body).Decode(t); err != nil {
		fmt.Println(err.Error())
	}
	o := t.Obu
	obu, _ := server.GetObu(o.ID, o.SPZ, o.Country, dbType)
	if obu == nil {
		w.Write([]byte("Not found"))
		// handle unexpexted OBU
		return
	}
	amount := processTicket(*t)
	server.SetTollAmount(obu, amount, dbType)
	obuJSON, _ := json.Marshal(obu)

	w.Write([]byte(obuJSON))
}

func obu_handler(w http.ResponseWriter, r *http.Request) {
	var o server.OnBoardUnit
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		fmt.Println(err.Error())
	}
	obu, _ := server.GetObu(o.ID, o.SPZ, o.Country, dbType)
	if obu == nil {
		w.Write([]byte("error: OBU not found"))
		return
	}
	if obu.Emission != o.Emission || obu.Weight != o.Weight || obu.Axles != o.Axles {
		server.UpdateObu(o.ID, o.SPZ, o.Country, o.Emission,
			fmt.Sprintf("%d", o.Weight), fmt.Sprintf("%d", o.Axles), dbType)
		w.Write([]byte("Modified parameters in OBU"))
	}
	byteJson, _ := json.Marshal(obu)
	w.Write(byteJson)

}

func geo_handler(w http.ResponseWriter, r *http.Request) {
	var result []byte
	var opt string = ""
	q, _ := url.PathUnescape(r.URL.RawQuery)
	if len(q) != 0 {
		m, _ := url.ParseQuery(q)
		js, _ := json.Marshal(m)

		var param *geoUrlParams
		json.Unmarshal(js, &param)
		if param.Version != nil {
			opt = "v"
		}

	}
	if opt == "v" {
		var versions []server.WptRecords
		server.LoadModel()
		for _, v := range server.Model {
			var section server.WptRecords
			section.Version = v.Version
			section.Name = v.Name
			section.Checksum = hash([]byte(fmt.Sprintf("%+v", v)))
			versions = append(versions, section)
		}
		result, _ = json.Marshal(versions)
	} else {
		server.LoadModel()
		result, _ = json.Marshal(server.Model)
	}

	w.Write([]byte(result))
}

func processTicket(t ticket) float64 {
	var distance float64 = 0.0
	var sazba float64 = 0.0
	var roadname string = ""
	var timestamp string = ""
	model := server.Model

	obu := t.Obu
	p := t.CheckPoints
	var i int = 0
	for ; i < len(p.I)-1; i++ {
		if p.I[i] == p.I[i+1] && server.IsDay(p.Time[i]) == server.IsDay(p.Time[i+1]) {
			//still the same paid road section, and still the same day or night
			lat1 := model[p.I[i]].LatRad[p.J[i]]
			lon1 := model[p.I[i]].LonRad[p.J[i]]
			lat2 := model[p.I[i+1]].LatRad[p.J[i+1]]
			lon2 := model[p.I[i+1]].LonRad[p.J[i+1]]
			distance += server.Haversine(lat1, lon1, lat2, lon2)
		} else if p.I[i] != p.I[i+1] || server.IsDay(p.Time[i]) != server.IsDay(p.Time[i+1]) {
			//end of the same paid road section, or changed from daytime to nightime and vice versa
			//For each road section there are different charge and for daytime and nightime
			roadname = model[p.I[i]].Name
			timestamp = p.Time[i]
			sazba += server.ExecSazba(distance, timestamp, obu.Weight,
				obu.Axles, obu.Category, obu.Emission, roadname)

			distance = 0.0
		}

	}

	roadname = model[p.I[i]].Name
	timestamp = p.Time[i]
	sazba += server.ExecSazba(distance, timestamp, obu.Weight,
		obu.Axles, obu.Category, obu.Emission, roadname)

	return sazba
}

func hash(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
