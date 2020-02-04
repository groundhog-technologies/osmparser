package element

import (
	"bytes"
	"encoding/gob"
	"github.com/paulmach/go.geojson"
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

	// Check type.
	var isMultiPolygon bool
	if v, ok := e.Relation.Tags["type"]; ok {
		if v == "multipolygon" {
			isMultiPolygon = true

		}
	}
	if isMultiPolygon {

		multiPolygon := [][][][]float64{}

		latLons := [][]float64{}
		polygon := [][][]float64{}
		pendingInnerPolygon := [][][]float64{}
		for _, emtMember := range e.Elements {
			var emtFeature *geojson.Feature
			switch emtMember.Type {
			case "Node":
				emtFeature = NodeElementToFeature(&emtMember)
			case "Way":
				emtFeature = WayElementToFeature(&emtMember)
			case "Relation":
				emtFeature = RelationElementToFeature(&emtMember)
			}

			// Get latLons from element's geojson feature.
			emtLatLons := [][]float64{}
			switch emtFeature.Geometry.Type {
			case geojson.GeometryPoint:
				emtLatLons = append(emtLatLons, emtFeature.Geometry.Point)
			case geojson.GeometryMultiPoint:
				emtLatLons = append(emtLatLons, emtFeature.Geometry.MultiPoint...)
			case geojson.GeometryLineString:
				emtLatLons = append(emtLatLons, emtFeature.Geometry.LineString...)
			case geojson.GeometryMultiLineString:
				for _, lineString := range emtFeature.Geometry.MultiLineString {
					emtLatLons = append(emtLatLons, lineString...)
				}
			case geojson.GeometryPolygon:
				multiPolygon = append(multiPolygon, emtFeature.Geometry.Polygon)
			case geojson.GeometryMultiPolygon:
				multiPolygon = append(multiPolygon, emtFeature.Geometry.MultiPolygon...)
			}
			var emtStartPoint, emtEndPoint []float64
			if len(emtLatLons) > 0 {
				emtStartPoint = emtLatLons[0]
				emtEndPoint = emtLatLons[len(emtLatLons)-1]
			}

			// Checkint the graft point.
			// AStart, AEnd, BStart, BEnd
			// Cases: (AEnd, BStart), (AEnd, BEnd), (AStart, BEnd), (AStart, BStart)
			if len(latLons) == 0 {
				latLons = append(latLons, emtLatLons...)
			} else {
				checkStartPoint := latLons[0]
				checkEndPoint := latLons[len(latLons)-1]
				if reflect.DeepEqual(checkEndPoint, emtStartPoint) {
					latLons = append(latLons, emtLatLons...)
				} else if reflect.DeepEqual(checkStartPoint, emtStartPoint) {
					newLatLons := [][]float64{}
					for i := len(emtLatLons) - 1; i >= 0; i-- {
						newLatLons = append(newLatLons, emtLatLons[i])
					}
					latLons = append(newLatLons, latLons...)
				} else if reflect.DeepEqual(checkEndPoint, emtEndPoint) {
					for i := len(emtLatLons) - 1; i >= 0; i-- {
						latLons = append(latLons, emtLatLons[i])
					}
				} else {
					// Equal to reflect.DeepEqual(checkEndPoint, emtEndPoint)
					latLons = append(emtLatLons, latLons...)
				}
			}

			// If area?
			// if latLons length > 1 && first point is equal to last point.
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
		}
		if len(polygon) > 0 {
			if len(pendingInnerPolygon) > 0 {
				polygon = append(polygon, pendingInnerPolygon...)
			}
			multiPolygon = append(multiPolygon, polygon)
		}
		f = geojson.NewMultiPolygonFeature(multiPolygon...)

	} else {
		// Sometime relation will missing type tag, so default we use CollectFeature.
		geometries := []*geojson.Geometry{}
		for _, emtMember := range e.Elements {
			switch emtMember.Type {
			case "Node":
				emtFeature := NodeElementToFeature(&emtMember)
				geometries = append(
					geometries,
					emtFeature.Geometry,
				)
			case "Way":
				emtFeature := WayElementToFeature(&emtMember)
				geometries = append(
					geometries,
					emtFeature.Geometry,
				)
			case "Relation":
				emtFeature := RelationElementToFeature(&emtMember)
				geometries = append(
					geometries,
					emtFeature.Geometry,
				)
			}
		}
		f = geojson.NewCollectionFeature(geometries...)
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
	return f
}
