package nominatium_test

import (
	"testing"

	"github.com/t0k4r/places/nominatium"
)

func TestOK(t *testing.T) {
	client, err := nominatium.New("github-com-t0k4r-places-test")
	if err != nil {
		t.Fatal(err)
	}
	places, err := client.Find("mosina")
	if err != nil {
		t.Fatal(err)
	}
	for _, place := range places {
		t.Logf("%+v\n", place)
	}
	places, err = client.Find("dębiec")
	if err != nil {
		t.Fatal(err)
	}
	for _, place := range places {
		t.Logf("%+v\n", place)
	}
}
