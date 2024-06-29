package nominatium_test

import (
	"testing"

	"github.com/t0k4r/places"
	"github.com/t0k4r/places/nominatium"
)

func TestOK(t *testing.T) {
	var client places.PlaceFinder
	var err error
	client, err = nominatium.NewClient("github-com-t0k4r-places-test")
	if err != nil {
		t.Fatal(err)
	}
	placesv, err := client.FromName("Poznań zamek")
	if err != nil {
		t.Fatal(err)
	}
	for i, place := range placesv {
		t.Logf("i:%v: name: '%v' lat:%v long:%v", i, place.Name, place.Lat, place.Long)
		plcs, err := client.FromLatLong(place.LatLong)
		if err != nil {
			t.Fatal(err)
		}
		for j, plc := range plcs {
			t.Logf("j:%v: name: '%v' lat:%v long:%v", j, plc.Name, plc.Lat, plc.Long)
		}
	}
	placesv, err = client.FromLatLong(places.LatLong{
		Lat:  21.37,
		Long: 69.42,
	})
	if err != nil {
		t.Fatal(err)
	}
	for i, place := range placesv {
		t.Logf("i:%v: name: '%v' lat:%v long:%v", i, place.Name, place.Lat, place.Long)
		plcs, err := client.FromName(place.Name)
		if err != nil {
			t.Fatal(err)
		}
		for j, plc := range plcs {
			t.Logf("j:%v: name: '%v' lat:%v long:%v", j, plc.Name, plc.Lat, plc.Long)
		}
	}

}
