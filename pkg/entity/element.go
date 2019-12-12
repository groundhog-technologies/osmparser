package entity

// LatLng .
type LatLng struct {
	Lat float64
	Lng float64
}

// Element .
type Element struct {
	ID         int64
	Type       string
	Points     []LatLng
	Tags       map[string]string
	FinalCodes []string
	NameCode   string
}
