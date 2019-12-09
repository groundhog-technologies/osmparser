package osm

import (
	"encoding/hex"
	"encoding/json"
	"hash/fnv"
	"strings"
)

// LatLng .
type LatLng struct {
	Lat float64
	Lng float64
}

// Element .
type Element struct {
	ID        int64
	Type      string
	Points    []LatLng
	Tags      map[string]string
	FinalCode string
	Cell      Cell
}

// Cell .
type Cell struct {
	Tags map[string]string
	Type string
}

// GenFinalCode .
// Turn tags to md5Hash.
func (e *Element) GenFinalCode() error {

	cell := Cell{
		Tags: make(map[string]string),
		Type: e.Type,
	}

	// Only take OSM Primary features tags.
	// https://wiki.openstreetmap.org/wiki/Map_Features
	primaryFeatures := []string{
		"aerialway",
		"aeroway",
		"amenity",
		"barrier",
		"boundary",
		"building",
		"craft",
		"emergency",
		"geological",
		"highway",
		"historic",
		"landuse",
		"leisure",
		"man_made",
		"military",
		"natural",
		"office",
		"place",
		"power",
		"public_transport",
		"railway",
		"route",
		"shop",
		"sport",
		"telecom",
		"tourism",
		"waterway",
	}

	for k, v := range e.Tags {
		for _, disableKey := range primaryFeatures {
			if strings.Contains(k, ":") {
				continue
			}
			if strings.Contains(k, disableKey) {
				switch k {
				case "amenity":
					cell.Tags[k] = v
				default:
					cell.Tags[k] = v

				}
			}
		}
	}
	b, err := json.Marshal(cell)
	if err != nil {
		return err
	}
	hasher := fnv.New32a()
	hasher.Write(b)
	e.FinalCode = hex.EncodeToString(hasher.Sum(nil))
	e.Cell = cell
	return nil
}
