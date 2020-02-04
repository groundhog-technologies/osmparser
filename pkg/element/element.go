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

func (e *Element) GetName() (string, bool) {
	var isName bool
	var name string
	var tags map[string]string
	switch e.Type {
	case "Node":
		tags = e.Node.Tags
	case "Way":
		tags = e.Node.Tags
	case "Relation":
		tags = e.Node.Tags
	default:
		tags = make(map[string]string)
	}
	if v, ok := tags["name"]; ok {
		name = v
		isName = true
	}
	return name, isName
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
	fc := geojson.NewFeatureCollection()
	f := RelationElementToFeature(e)
	fc.AddFeature(f)
	rawJSON, _ := fc.MarshalJSON()
	return rawJSON
}

func (e *Element) IsArea() bool {
	// Define is way is area(polygon) or not.
	// https://wiki.openstreetmap.org/wiki/Key:area
	var isPolygon bool
	if val, ok := e.Way.Tags["area"]; ok && val == "yes" {
		// Is closedPolylines.
		// https://wiki.openstreetmap.org/wiki/Way#Closed_way
		if _, ok := e.Way.Tags["highway"]; ok {
		} else if _, ok := e.Way.Tags["barrier"]; ok {
		} else {
			isPolygon = true
		}
	}
	return isPolygon
}
