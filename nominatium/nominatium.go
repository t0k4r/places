package nominatium

import (
	"database/sql"
	"encoding/json"
	"log/slog"
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

func NewClient(userAgent string) (Client, error) {
	client := Client{userAgent: userAgent}
	var err error
	client.db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		return client, err
	}
	for _, table := range strings.Split(schema, ";") {
		_, err := client.db.Exec(table)
		return client, err
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
func (c *Client) netFromName(name string) ([]places.Place, error) {
	lock()
	defer unlock()
	return nil, nil
}
func (c *Client) dbFromName(name string) ([]places.Place, error) {
	var plcs []places.Place
	rows, err := c.db.Query("select data from places where name = ?", name)
	if err != nil {
		return plcs, err
	}
	if rows.Next() {
		var buf string
		err = rows.Scan(&buf)
		if err != nil {
			return plcs, err
		}
		err = json.Unmarshal([]byte(buf), &plcs)
		if err != nil {
			return plcs, err
		}
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
	return nil, nil
}
func (c *Client) dbFromLatLong(latLong places.LatLong) ([]places.Place, error) {
	var plcs []places.Place
	rows, err := c.db.Query("select data from places where lat = ? and long = ?", latLong.Lat, latLong.Long)
	if err != nil {
		return plcs, err
	}
	if rows.Next() {
		var buf string
		err = rows.Scan(&buf)
		if err != nil {
			return plcs, err
		}
		err = json.Unmarshal([]byte(buf), &plcs)
		if err != nil {
			return plcs, err
		}
	}
	return plcs, nil
}
