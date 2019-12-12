package classifier

import (
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"github.com/sirupsen/logrus"
	"github.com/uber/h3-go"
	"osm-parser/pkg/entity"
	"osm-parser/pkg/mapfeature"
	"reflect"
	"sort"
)

// AreaClassifier .
type AreaClassifier struct {
	Elements    []entity.Element
	Resolution  int
	MapFeatures mapfeature.MapFeatures
	GeoPolygon  h3.GeoPolygon
}

// H3Cell .
type H3Cell struct {
	H3Index h3.H3Index
	POIs    map[string]int
}

// CellToClustersCoordinates .
func (c *H3Cell) CellToClustersCoordinates(finalCodes []string) clusters.Coordinates {
	coordinates := clusters.Coordinates{}
	for _, code := range finalCodes {
		coordinates = append(coordinates, float64(c.POIs[code]))
	}
	return coordinates
}

// Run .
func (c *AreaClassifier) Run() error {
	h3Idxs := h3.Polyfill(c.GeoPolygon, c.Resolution)

	// Filter element not in polygon.
	newElements := []entity.Element{}
	h3IdxsMap := make(map[h3.H3Index]int)
	for _, idx := range h3Idxs {
		h3IdxsMap[idx] = 1
	}
	for _, element := range c.Elements {
		geoFence := []h3.GeoCoord{}
		elementH3Idxs := []h3.H3Index{}
		for _, latLng := range element.Points {
			h3Idx := h3.FromGeo(
				h3.GeoCoord{
					Latitude:  latLng.Lat,
					Longitude: latLng.Lng,
				},
				c.Resolution,
			)
			elementH3Idxs = append(h3Idxs, h3Idx)
			geoFence = append(
				geoFence,
				h3.GeoCoord{
					Latitude:  latLng.Lat,
					Longitude: latLng.Lng,
				},
			)
		}
		inElementH3Idxs := h3.Polyfill(h3.GeoPolygon{Geofence: geoFence}, c.Resolution)
		for _, h3Idx := range inElementH3Idxs {
			elementH3Idxs = append(elementH3Idxs, h3Idx)
		}

		isInPolygon := true
		for _, idx := range elementH3Idxs {
			if _, ok := h3IdxsMap[idx]; !ok {
				isInPolygon = false
				break
			}
		}
		if isInPolygon {
			newElements = append(newElements, element)
			logrus.Debugf("%#v in polygon", element)
		}
	}
	c.Elements = newElements
	h3IdxsMap = nil

	// Gen poiTypeMap
	poiTypeMap := make(map[string]int)
	for _, element := range c.Elements {
		for _, code := range element.FinalCodes {
			poiTypeMap[code] = 1
		}
	}
	poiFinalCodes := []string{}
	for k := range poiTypeMap {
		poiFinalCodes = append(poiFinalCodes, k)
	}
	sort.Strings(poiFinalCodes)
	logrus.Info("Finish gen poi finalCode arr")

	cellStatistics := make(map[h3.H3Index]H3Cell)
	for _, h3Idx := range h3Idxs {
		cellPOIs := make(map[string]int)
		for _, fc := range poiFinalCodes {
			cellPOIs[fc] = 0
		}
		cellStatistics[h3Idx] = H3Cell{
			H3Index: h3Idx,
			POIs:    cellPOIs,
		}
	}
	logrus.Info("Finish gen cellStatistics")

	for _, element := range c.Elements {
		logrus.Debug(element)
		geoFence := []h3.GeoCoord{}
		elementH3Idxs := []h3.H3Index{}
		for _, latLng := range element.Points {
			h3Idx := h3.FromGeo(
				h3.GeoCoord{
					Latitude:  latLng.Lat,
					Longitude: latLng.Lng,
				},
				c.Resolution,
			)
			elementH3Idxs = append(h3Idxs, h3Idx)
			geoFence = append(
				geoFence,
				h3.GeoCoord{
					Latitude:  latLng.Lat,
					Longitude: latLng.Lng,
				},
			)
		}
		inElementH3Idxs := h3.Polyfill(h3.GeoPolygon{Geofence: geoFence}, c.Resolution)
		for _, h3Idx := range inElementH3Idxs {
			elementH3Idxs = append(elementH3Idxs, h3Idx)
		}

		for _, h3Idx := range elementH3Idxs {
			if _, ok := cellStatistics[h3Idx]; ok {
				for _, code := range element.FinalCodes {
					cellStatistics[h3Idx].POIs[code]++
				}
			}
		}
	}

	logrus.Info("Start kmeans")
	d := clusters.Observations{}
	for _, cell := range cellStatistics {
		d = append(d, cell.CellToClustersCoordinates(poiFinalCodes))
	}

	km := kmeans.New()
	clusters, err := km.Partition(d, 12)
	if err != nil {
		return err
	}
	for idx, c := range clusters {
		for k, v := range c.Center {
			finalCode := poiFinalCodes[k]
			logrus.Infof("%v:%v: %v\n", idx, finalCode, v)
		}

		num := 0
		for _, observation := range c.Observations {
			if num > 10 {
				break
			}
			for _, cell := range cellStatistics {
				if reflect.DeepEqual(cell.CellToClustersCoordinates(poiFinalCodes), observation) {
					num++
					logrus.Debug(h3.ToGeo(cell.H3Index))
					if num > 10 {
						break
					}
				}
			}
		}

	}
	return nil
}
