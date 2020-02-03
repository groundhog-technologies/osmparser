package element

import (
	"bytes"
	"encoding/gob"
	"github.com/paulmach/go.geojson"
	"github.com/thomersch/gosmparse"
)

// Element .
type Element struct {
	Type     string // Node, Way, Relation
	Node     gosmparse.Node
	Way      gosmparse.Way
	Role     string
	Relation gosmparse.Relation
	Elements []Element
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

// ToJSON .
func (e *Element) ToJSON() []byte {
	var rawJSON []byte
	switch e.Type {
	case "Node":
		rawJSON = e.nodeToJSON()
	case "Way":
		rawJSON = e.wayToJSON()
	case "Relation":
		rawJSON = e.relationToJSON()
	}
	return rawJSON
}

// nodeToJSON converts node element to json.
func (e *Element) nodeToJSON() []byte {
	fc := geojson.NewFeatureCollection()
	fc.AddFeature(NodeElementToFeature(e))
	rawJSON, _ := fc.MarshalJSON()
	return rawJSON
}

// wayToJSON converts way element to JSON.
func (e *Element) wayToJSON() []byte {
	fc := geojson.NewFeatureCollection()
	fc.AddFeature(WayElementToFeature(e))
	rawJSON, _ := fc.MarshalJSON()
	return rawJSON
}

// relationToJSON converts relation element to JSON.
func (e *Element) relationToJSON() []byte {
	return []byte{}
}
