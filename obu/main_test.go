package main

import (
	"fmt"
	"testing"
)

func TestFindPair(t *testing.T) {
	tests := []struct {
		array1 []int
		array2 []int
		value1 int
		value2 int
		exp    int
	}{
		{[]int{1, 2, 4, 6, 10}, []int{3, 2, 5, 7, 9}, 2, 2, 1},
	}
	for _, test := range tests {
		if got := findPair(test.array1, test.array2, test.value1, test.value2); got != test.exp {
			t.Errorf("At input %v, %v, %d, %d \nexpected '%d', but got '%d'", test.array1, test.array2, test.value1, test.value2, test.exp, got)
		}
	}
}

func TestFindValue(t *testing.T) {
	tests := []struct {
		array []int
		value int
		exp   int
	}{
		{[]int{1, 2, 4, 6, 10}, 2, 1},
		{[]int{1, 2, 4, 6, 10}, 7, -1},
	}
	for _, test := range tests {
		if got := findValue(test.array, test.value); got != test.exp {
			t.Errorf("At input %v, %d \nexpected '%d', but got '%d'", test.array, test.value, test.exp, got)
		}
	}

}

func TestHaversine(t *testing.T) {
	tests := []struct {
		lat1 float64
		lon1 float64
		lat2 float64
		lon2 float64
		exp  float64
	}{
		{0.8833989070424129, 0.2637513989411649,
			0.8833961319689024, 0.2637535456961449, 19.69511},
	}

	for _, test := range tests {
		if got := haversine(test.lat1, test.lon1, test.lat2, test.lon2); fmt.Sprintf("%.2f", got) != fmt.Sprintf("%.2f", test.exp) {
			t.Errorf(`at input lat1 '%.5f', lon1 '%.5f'
			lat2 '%.5f', lon2 '%.5f'
			expected '%.5f', but got '%.5f'`, test.lat1, test.lon1, test.lat2, test.lon2, test.exp, got)
		}
	}

}

func TestDegreesToRadians(t *testing.T) {
	tests := []struct {
		value float64
		exp   string
	}{
		{30, "0.52360"},
		{90, "1.57080"},
		{57.2957795, "1.00000"},
	}
	for _, test := range tests {
		if got := degreesToRadians(test.value); fmt.Sprintf("%.5f", got) != test.exp {
			t.Errorf("at input '%.5f' expected '%s', but got '%.5f'", test.value, test.exp, got)
		}
	}
}
