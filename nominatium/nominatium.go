package nominatium

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "embed"

	_ "github.com/glebarez/go-sqlite"
	"github.com/t0k4r/places"
)

//go:embed schema.sql
var schema string

var mut sync.Mutex = sync.Mutex{}

func lock() {
	mut.Lock()
	time.Sleep(time.Second / 2)
}
func unlock() {
	time.Sleep(time.Second / 2)
	mut.Unlock()
}

type Client struct {
	userAgent string
	db        *sql.DB
}

type netResp struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Long        string `json:"lon"`
}

func NewClient(userAgent string) (places.PlaceFinder, error) {
	client := &Client{userAgent: userAgent}
	var err error
	client.db, err = sql.Open("sqlite", "nominatium.sqlite")
	if err != nil {
		return client, err
	}
	for _, table := range strings.Split(schema, ";") {
		_, err := client.db.Exec(table)
		if err != nil {
			return client, err
		}
	}
	return client, nil
}

func (c *Client) FromName(name string) ([]places.Place, error) {
	places, err := c.dbFromName(name)
	if err != nil {
		slog.Warn(err.Error())
	} else if len(places) != 0 {
		return places, err
	}
	places, err = c.netFromName(name)
	return places, err
}

func (c *Client) readNetRespMany(body io.Reader, query string) ([]places.Place, error) {
	var jarr []netResp
	err := json.NewDecoder(body).Decode(&jarr)
	if err != nil {
		return nil, err
	}
	var plcs []places.Place
	for _, j := range jarr {
		ll := places.LatLong{}
		ll.Lat, err = strconv.ParseFloat(j.Lat, 64)
		if err != nil {
			return plcs, err
		}
		ll.Long, err = strconv.ParseFloat(j.Long, 64)
		if err != nil {
			return plcs, err
		}
		plc := places.Place{
			LatLong: ll,
			Name:    j.DisplayName,
		}
		plcs = append(plcs, plc)
		c.dbInsert(query, plc)
	}
	return plcs, nil
}

func (c *Client) readNetRespOne(body io.Reader, query string) ([]places.Place, error) {
	var jarr netResp
	err := json.NewDecoder(body).Decode(&jarr)
	if err != nil {
		return nil, err
	}

	ll := places.LatLong{}
	ll.Lat, err = strconv.ParseFloat(jarr.Lat, 64)
	if err != nil {
		return nil, err
	}
	ll.Long, err = strconv.ParseFloat(jarr.Long, 64)
	if err != nil {
		return nil, err
	}
	plc := places.Place{
		LatLong: ll,
		Name:    jarr.DisplayName,
	}
	c.dbInsert(query, plc)

	return []places.Place{plc}, nil
}

func (c *Client) netFromName(name string) ([]places.Place, error) {
	lock()
	defer unlock()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://nominatim.openstreetmap.org/search.php?q=%v&format=jsonv2", url.QueryEscape(name)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return c.readNetRespMany(resp.Body, name)
}

func (c *Client) dbInsert(query string, place places.Place) error {
	_, err := c.db.Exec(`insert into places("query", name, lat, long) values ($1, $2, $3, $4)`, query, place.Name, place.Lat, place.Long)
	return err
}

func (c *Client) dbFromName(name string) ([]places.Place, error) {
	var plcs []places.Place
	rows, err := c.db.Query("select distinct name, lat, long from places where name = $1 or query = $1", name)
	if err != nil {
		return plcs, err
	}
	for rows.Next() {
		var plc places.Place
		err = rows.Scan(&plc.Name, &plc.Lat, &plc.Long)
		if err != nil {
			return plcs, err
		}
		plcs = append(plcs, plc)
	}
	return plcs, nil
}

func (c *Client) FromLatLong(latLong places.LatLong) ([]places.Place, error) {
	places, err := c.dbFromLatLong(latLong)
	if err != nil {
		slog.Warn(err.Error())
	} else if len(places) != 0 {
		return places, err
	}
	places, err = c.netFromLatLong(latLong)
	return places, err
}

func (c *Client) netFromLatLong(latLong places.LatLong) ([]places.Place, error) {
	lock()
	defer unlock()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://nominatim.openstreetmap.org/reverse.php?lat=%v&lon=%v&format=jsonv2", latLong.Lat, latLong.Long), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return c.readNetRespOne(resp.Body, "")
}
func (c *Client) dbFromLatLong(latLong places.LatLong) ([]places.Place, error) {
	var plcs []places.Place
	rows, err := c.db.Query("select distinct name, lat, long  from places where lat = $1 and long = $2", latLong.Lat, latLong.Long)
	if err != nil {
		return plcs, err
	}
	for rows.Next() {
		var plc places.Place
		err = rows.Scan(&plc.Name, &plc.Lat, &plc.Long)
		if err != nil {
			return plcs, err
		}
		plcs = append(plcs, plc)
	}
	return plcs, nil
}
