package element

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/paulmach/go.geojson"
	"github.com/sirupsen/logrus"
	"reflect"
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
	isArea := e.IsArea()

	if isArea {
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

	// logrus.Infof("%+v", e.Relation)
	var isMultiPolygon bool
	if v, ok := e.Relation.Tags["type"]; ok {
		switch v {
		case "multipolygon":
			isMultiPolygon = true
			multiPolygon := [][][][]float64{}

			latLons := [][]float64{}
			polygon := [][][]float64{}
			pendingInnerPolygon := [][][]float64{}
			for _, emtMember := range e.Elements {
				var newF *geojson.Feature
				switch emtMember.Type {
				case "Node":
					newF = NodeElementToFeature(&emtMember)
				case "Way":
					newF = WayElementToFeature(&emtMember)
				case "Relation":
					newF = RelationElementToFeature(&emtMember)
				}

				switch newF.Geometry.Type {
				case geojson.GeometryPoint:
					latLons = append(latLons, newF.Geometry.Point)
				case geojson.GeometryMultiPoint:
					latLons = append(latLons, newF.Geometry.MultiPoint...)
				case geojson.GeometryLineString:
					latLons = append(latLons, newF.Geometry.LineString...)
				case geojson.GeometryMultiLineString:
					for _, lineString := range newF.Geometry.MultiLineString {
						latLons = append(latLons, lineString...)
					}
				case geojson.GeometryPolygon:
					multiPolygon = append(multiPolygon, newF.Geometry.Polygon)
				case geojson.GeometryMultiPolygon:
					multiPolygon = append(multiPolygon, newF.Geometry.MultiPolygon...)
				}

				// If area?
				if len(latLons) > 1 && reflect.DeepEqual(latLons[0], latLons[len(latLons)-1]) {
					switch emtMember.Role {
					case "outer":
						if len(polygon) > 0 {
							multiPolygon = append(multiPolygon, polygon)
							polygon = [][][]float64{}
						}
						polygon = append(
							polygon,
							latLons,
						)
						if len(pendingInnerPolygon) > 0 {
							polygon = append(polygon, pendingInnerPolygon...)
							pendingInnerPolygon = [][][]float64{}
						}
					case "inner":
						if len(polygon) == 0 {
							pendingInnerPolygon = append(
								pendingInnerPolygon,
								latLons,
							)
						} else {
							polygon = append(
								polygon,
								latLons,
							)
						}
					}
					latLons = [][]float64{}
				}
				logrus.Infof("%v: %+v %+v", emtMember.Role, newF, newF.Geometry)
			}
			if len(polygon) > 0 {
				if len(pendingInnerPolygon) > 0 {
					polygon = append(polygon, pendingInnerPolygon...)
				}
				multiPolygon = append(multiPolygon, polygon)
			}
			f = geojson.NewMultiPolygonFeature(multiPolygon...)
		default: // not area
			geometries := []*geojson.Geometry{}
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
			f = geojson.NewCollectionFeature(geometries...)
		}
	}
	// Add tag to property.

	relID := "relation" + "/" + strconv.FormatInt(e.Relation.ID, 10)
	f.ID = relID
	f.SetProperty("osmid", relID)
	f.SetProperty("osmType", "relation")
	for k, v := range e.Relation.Tags {
		f.SetProperty(
			k, v,
		)
	}
	if isMultiPolygon {
		fc := geojson.NewFeatureCollection()
		fc.AddFeature(f)
		rawJSON, _ := fc.MarshalJSON()
		fmt.Println(string(rawJSON) + "\n----------------------------\n")
	}
	return f
}
