package places

type LatLong struct {
	Lat  float64
	Long float64
}

type Place struct {
	LatLong
	Name string
}

type PlaceFinder interface {
	FromName(string) ([]Place, error)
	FromLatLong(LatLong) ([]Place, error)
}
