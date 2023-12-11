package server

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type weight struct {
	Axles2 float64 `json:"2"`
	Axles3 float64 `json:"3"`
	Axles4 float64 `json:"4"`
	Axles5 float64 `json:"5"`
}

type emission struct {
	W3575 weight `json:"35-75"`
	W7512 weight `json:"75-12"`
	W12   weight `json:"12"`
}

type vehicle struct {
	E04   emission `json:"0-4"`
	E5EEV emission `json:"5-EEV"`
	E6    emission `json:"6"`
	Ecng  emission `json:"CNG"`
}

type Sazba struct {
	N  vehicle `json:"N"`
	M2 vehicle `json:"M2"`
	// m3 vehicle `json:"M3"`
}

var Ratio float64 = 100.0 // pay for each 100 meters
const DIR = "sazba"

var dDayFilename string = DIR + "/d-day.json"
var dNightFilename string = DIR + "/d-night.json"
var iDayFilename string = DIR + "/i-day.json"
var iNightFilename string = DIR + "/i-night.json"

var dDay Sazba
var dNight Sazba
var iDay Sazba
var iNight Sazba

func LoadSazba() {
	load(dDayFilename, &dDay)
	load(dNightFilename, &dNight)
	load(iDayFilename, &iDay)
	load(iNightFilename, &iNight)
	//	c := Charge(100, time.Now().Format(time.RFC3339), 8500, 4, "M3", "4", "I35")

}

// Compute a charge by a distance for using a toll road
// distance in meters
func ExecSazba(distance float64, timedate string, weightKilo int,
	numberaxles int, category string, emissionCategory string, roadname string) float64 {
	var result float64 = 0.0
	if distance < Ratio {
		return result
	}

	var s Sazba
	whichSazba(roadname, timedate, &s)
	var v vehicle
	whichCategory(category, s, &v)
	var e emission
	whichEmission(emissionCategory, v, &e)
	var w weight
	whichWeight(weightKilo, e, &w)
	var charge float64
	whichAxles(numberaxles, category, w, &charge)
	d := distance / Ratio

	result = charge * d
	return result
}

func load(filename string, sazba *Sazba) {
	jsonFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	json.Unmarshal(byteValue, sazba)

}

// Distinguish I. class road or highway
func whichSazba(roadname string, timedate string, s *Sazba) {
	if IsDay(timedate) {
		if found := strings.HasPrefix(roadname, "D"); found {
			*s = dDay
		} else if found := strings.HasPrefix(roadname, "I"); found {
			*s = iDay
		} else {
			fmt.Printf("Roadname %s does not belong to any type of road.", roadname)
		}

	} else {
		// Night charge
		if found := strings.HasPrefix(roadname, "D"); found {
			*s = dNight
		} else if found := strings.HasPrefix(roadname, "I"); found {
			*s = iNight
		} else {
			fmt.Printf("Roadname %s does not belong to any type of road.", roadname)
		}

	}
}

func whichCategory(category string, s Sazba, v *vehicle) {
	switch category {
	case "N":
		*v = s.N
	case "M2", "M3":
		*v = s.M2
	default:
		fmt.Printf("Category %s does not exist.", category)
	}

}

func whichEmission(emissionStr string, v vehicle, e *emission) {
	switch emissionStr {
	case "0", "1", "2", "3", "4", "euro0":
		*e = v.E04
	case "5", "EEV":
		*e = v.E5EEV
	case "6":
		*e = v.E6
	case "CNG":
		*e = v.Ecng
	default:
		fmt.Printf("Emission %s does not exist.", emissionStr)

	}

}

func whichWeight(weightKilo int, e emission, w *weight) {
	//Weight in kilograms

	if weightKilo > 3500 && weightKilo < 7500 {
		*w = e.W3575
	} else if weightKilo >= 7500 && weightKilo < 12000 {
		*w = e.W7512
	} else if weightKilo >= 12000 {
		*w = e.W12
	} else {
		fmt.Printf("This %d weight has no place in database.", weightKilo)
	}
}

func whichAxles(numberaxles int, category string, w weight, a *float64) {
	if numberaxles > 5 && category == "N" {
		numberaxles = 5
	} else if numberaxles > 3 && (category == "M2" || category == "M3") {
		numberaxles = 3
	}

	switch numberaxles {
	case 2:
		*a = w.Axles2
	case 3:
		*a = w.Axles3
	case 4:
		*a = w.Axles4
	case 5:
		*a = w.Axles5
	default:
		fmt.Printf("This %d number of axles has no place in database.", numberaxles)
	}
}

func IsDay(timedate string) bool {
	t, err := time.Parse(time.RFC3339, timedate)
	if err != nil {
		fmt.Println(err)
	}
	if t.Hour() > 21 || t.Hour() <= 4 {
		return false
	}
	return true
}
