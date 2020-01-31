package osm

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
	}
	return rawJSON
}

// nodeToJSON transform node element to json.
func (e *Element) nodeToJSON() []byte {
	fc := geojson.NewFeatureCollection()
	f := geojson.NewPointFeature(
		[]float64{e.Node.Lon, e.Node.Lat},
	)

	nodeID := "node" + "/" + strconv.FormatInt(e.Node.ID, 10)
	f.ID = nodeID
	f.SetProperty("osmid", nodeID)
	f.SetProperty("osmType", "Node")

	// Add tag to property.
	for k, v := range e.Node.Tags {
		f.SetProperty(
			k, v,
		)
	}
	fc.AddFeature(f)
	rawJSON, _ := fc.MarshalJSON()
	return rawJSON
}

// wayToJSON .
func (e *Element) wayToJSON() []byte {
	return []byte{}
}
