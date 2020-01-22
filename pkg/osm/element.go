package osm

import (
	"github.com/paulmach/go.geojson"
	"github.com/thomersch/gosmparse"
)

// Element .
type Element struct {
	Type     string
	Node     gosmparse.Node
	Way      gosmparse.Way
	Relation gosmparse.Relation
}

// ToJSON .
func (e *Element) ToJSON() []byte {
	var rawJSON []byte
	switch e.Type {
	case "Node":
		g := geojson.NewPointGeometry(
			[]float64{e.Node.Lon, e.Node.Lat},
		)
		rawJSON, _ = g.MarshalJSON()
	}
	return rawJSON
}

type jsonNode struct {
	ID   int64             `json:"id"`
	Type string            `json:"type"`
	Lat  float64           `json:"lat"`
	Lon  float64           `json:"lon"`
	Tags map[string]string `json:"tags"`
}

type jsonWay struct {
	ID   int64             `json:"id"`
	Type string            `json:"type"`
	Tags map[string]string `json:"tags"`
	// NodeIDs   []int64             `json:"refs"`
	Centroid map[string]string   `json:"centroid"`
	Bounds   map[string]string   `json:"bounds"`
	Nodes    []map[string]string `json:"nodes,omitempty"`
}

type jsonRelation struct {
	ID       int64             `json:"id"`
	Type     string            `json:"type"`
	Tags     map[string]string `json:"tags"`
	Centroid map[string]string `json:"centroid"`
	Bounds   map[string]string `json:"bounds"`
}
