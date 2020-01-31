package osm

import (
	"bytes"
	"encoding/gob"
	"github.com/paulmach/go.geojson"
	"github.com/thomersch/gosmparse"
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
	}
	return rawJSON
}

// nodeToJSON transform node element to json.
func (e *Element) nodeToJSON() []byte {
	fc := geojson.NewFeatureCollection()
	pf := geojson.NewPointFeature(
		[]float64{e.Node.Lon, e.Node.Lat},
	)

	pf.SetProperty("osmid", e.Node.ID)
	pf.SetProperty("osmType", "Node")

	// Add tag to property.
	for k, v := range e.Node.Tags {
		pf.SetProperty(
			k, v,
		)
	}
	fc.AddFeature(pf)
	rawJSON, _ := fc.MarshalJSON()
	return rawJSON
}

// wayToJSON .
func (e *Element) wayToJSON() []byte {
	return []byte{}
}
