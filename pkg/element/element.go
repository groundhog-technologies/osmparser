package element

import (
	"bytes"
	"encoding/gob"
	"github.com/paulmach/go.geojson"
	"github.com/thomersch/gosmparse"
	"strconv"
)

// ByteToElement transform byte to element.
func ByteToElement(byteArr []byte) (Element, error) {
	decoder := gob.NewDecoder(bytes.NewReader(byteArr))
	var element Element
	err := decoder.Decode(&element)
	return element, err
}

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

func nodeToFeature(e *Element) *geojson.Feature {
	f := geojson.NewPointFeature(
		[]float64{e.Node.Lon, e.Node.Lat},
	)

	nodeID := "node" + "/" + strconv.FormatInt(e.Node.ID, 10)
	f.ID = nodeID
	f.SetProperty("osmid", nodeID)
	f.SetProperty("osmType", "node")

	// Add tag to property.
	for k, v := range e.Node.Tags {
		f.SetProperty(
			k, v,
		)
	}
	return f
}

// nodeToJSON converts node element to json.
func (e *Element) nodeToJSON() []byte {
	fc := geojson.NewFeatureCollection()
	f := nodeToFeature(e)
	fc.AddFeature(f)
	rawJSON, _ := fc.MarshalJSON()
	return rawJSON
}

// wayToJSON converts way element to JSON.
func (e *Element) wayToJSON() []byte {

	// collect latlon
	latLngs := [][]float64{}
	for _, member := range e.Elements {
		latLngs = append(
			latLngs,
			[]float64{member.Node.Lon, member.Node.Lat},
		)
	}

	fc := geojson.NewFeatureCollection()
	var f *geojson.Feature

	// Define is way is polygon or not.
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

	if isPolygon {
		f = geojson.NewPolygonFeature([][][]float64{latLngs})
	} else {
		f = geojson.NewLineStringFeature(latLngs)
	}

	wayID := "way" + "/" + strconv.FormatInt(e.Way.ID, 10)
	f.ID = wayID
	f.SetProperty("osmid", wayID)
	f.SetProperty("osmType", "way")

	// Add tag to property.
	for k, v := range e.Way.Tags {
		f.SetProperty(
			k, v,
		)
	}
	fc.AddFeature(f)

	rawJSON, _ := fc.MarshalJSON()
	return rawJSON
}

// relationToJSON converts relation element to JSON.
func (e *Element) relationToJSON() []byte {
	return []byte{}
}
