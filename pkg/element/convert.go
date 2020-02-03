package element

import (
	"bytes"
	"encoding/gob"
	"github.com/paulmach/go.geojson"
	// "github.com/sirupsen/logrus"
	"strconv"
)

// ByteToElement transform byte to element.
func ByteToElement(byteArr []byte) (Element, error) {
	decoder := gob.NewDecoder(bytes.NewReader(byteArr))
	var element Element
	err := decoder.Decode(&element)
	return element, err
}

func NodeElementToFeature(e *Element) *geojson.Feature {
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

func WayElementToFeature(e *Element) *geojson.Feature {
	// collect latlon
	latLngs := [][]float64{}
	for _, member := range e.Elements {
		latLngs = append(
			latLngs,
			[]float64{member.Node.Lon, member.Node.Lat},
		)
	}

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
	return f
}

func RelationElementToFeature(e *Element) *geojson.Feature {
	var f *geojson.Feature

	geometries := []*geojson.Geometry{}
	// logrus.Infof("%+v", e.Relation)
	if v, ok := e.Relation.Tags["type"]; ok {
		switch v {
		case "multipolygon":
		default: // not area
			for _, emtMember := range e.Elements {
				switch emtMember.Type {
				case "Node":
					newF := NodeElementToFeature(&emtMember)
					geometries = append(
						geometries,
						newF.Geometry,
					)
				case "Way":
					newF := WayElementToFeature(&emtMember)
					geometries = append(
						geometries,
						newF.Geometry,
					)
				case "Relation":
					newF := RelationElementToFeature(&emtMember)
					geometries = append(
						geometries,
						newF.Geometry,
					)
				}
			}
		}
	}
	f = geojson.NewCollectionFeature(geometries...)
	// Add tag to property.
	for k, v := range e.Relation.Tags {
		f.SetProperty(
			k, v,
		)
	}
	return f
}
