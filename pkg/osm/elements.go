package osm

// Element .
type Element struct {
	ID     int64
	Type   string
	Points []LatLng
	Tags   map[string]string
}

// LatLng .
type LatLng struct {
	Lat float64
	Lng float64
}
