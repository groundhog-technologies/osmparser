package osm

import (
	"bytes"
	"encoding/gob"
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

// ToByte .
func (e *Element) ToByte() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(e)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

// Transform byte to element.
func ByteToElement(byteArr []byte) (*Element, error) {
	decoder := gob.NewDecoder(bytes.NewReader(byteArr))
	var element Element
	err := decoder.Decode(&element)
	return &element, err
}
