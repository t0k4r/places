package places

type PlaceType string

const (
	Other         PlaceType = "other"
	Neighbourhood           = "neighbourhood"
	Suburb                  = "suburb"
	City                    = "city"
	Town                    = "town"
	Village                 = "village"
	Municipality            = "municipality"
	County                  = "county"
	State                   = "state"
	Country                 = "country"
)

type Place struct {
	Type          PlaceType
	Lat           float64
	Lon           float64
	Name          string
	Neighbourhood string
	Suburb        string
	City          string
	Town          string
	Village       string
	Municipality  string
	County        string
	State         string
	Country       string
}

type Finder interface {
	Find(name string) ([]Place, error)
}
