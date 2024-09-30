package nominatium

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/t0k4r/places"
	_ "modernc.org/sqlite"
)

var schema string = `
CREATE TABLE IF NOT EXISTS places (
	"query" TEXT,
    "json" TEXT
)
`

var mut sync.Mutex = sync.Mutex{}

func lock() {
	mut.Lock()
	time.Sleep(time.Second / 2)
}
func unlock() {
	time.Sleep(time.Second / 2)
	mut.Unlock()
}

type apiResp struct {
	Lat         float64 `json:"lat,string"`
	Lon         float64 `json:"lon,string"`
	AddressType string  `json:"addresstype"`
	DisplayName string  `json:"display_name"`
	Address     struct {
		Neighbourhood string `json:"neighbourhood"`
		Suburb        string `json:"suburb"`
		Village       string `json:"village"`
		Town          string `json:"town"`
		City          string `json:"city"`
		Municipality  string `json:"municipality"`
		County        string `json:"county"`
		State         string `json:"state"`
		Country       string `json:"country"`
	} `json:"address"`
}

func (resp *apiResp) placeType() places.PlaceType {
	switch resp.AddressType {
	case "neighbourhood":
		return places.Neighbourhood
	case "suburb":
		return places.Suburb
	case "city":
		return places.City
	case "village":
		return places.Village
	case "town":
		return places.Town
	case "municipality":
		return places.Municipality
	case "county":
		return places.County
	case "state":
		return places.State
	case "country":
		return places.Country
	}
	return places.Other
}

type client struct {
	userAgent string
	db        *sql.DB
}

func New(userAgent string) (f places.Finder, err error) {
	c := client{userAgent: userAgent}
	c.db, err = sql.Open("sqlite", "nominatium.sqlite")
	if err != nil {
		return &c, err
	}
	_, err = c.db.Exec(schema)
	return &c, err
}
func (c *client) Find(name string) (results []places.Place, err error) {
	results, err = dbQuery(c.db, name)
	if len(results) != 0 || (err != sql.ErrNoRows && err != nil) {
		return results, err
	}
	lock()
	defer unlock()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://nominatim.openstreetmap.org/search.php?q=%v&format=jsonv2&addressdetails=1", url.QueryEscape(name)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return results, err
	}
	var resps []apiResp
	err = json.NewDecoder(r.Body).Decode(&resps)
	if err != nil {
		return results, err
	}
	for _, resp := range resps {
		result := places.Place{
			Type:          resp.placeType(),
			Lat:           resp.Lat,
			Lon:           resp.Lon,
			Name:          resp.DisplayName,
			Neighbourhood: resp.Address.Neighbourhood,
			Suburb:        resp.Address.Suburb,
			City:          resp.Address.City,
			Town:          resp.Address.Town,
			Village:       resp.Address.Village,
			Municipality:  resp.Address.Municipality,
			County:        resp.Address.County,
			State:         resp.Address.State,
			Country:       resp.Address.Country,
		}
		results = append(results, result)
	}
	return results, nil
}
func dbQuery(db *sql.DB, query string) (results []places.Place, err error) {
	rows, err := db.Query("select * from places where query=$1", query)
	if err != nil {
		return results, err
	}
	for rows.Next() {
		var result places.Place
		var j string
		if err := rows.Scan(&j); err != nil {
			return results, nil
		}
		if err = json.Unmarshal([]byte(j), &result); err != nil {
			return results, nil
		}
		results = append(results, result)
	}
	return results, nil
}
func dbInsert(db *sql.DB, query string, resp apiResp) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	_, err = db.Exec("insert into places(query, json) values ($1, $2)", query, string(b))
	return err
}
