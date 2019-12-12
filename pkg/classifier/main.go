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

	h3Idxs := h3.Polyfill(c.GeoPolygon, c.Resolution)
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

	for _, element := range c.Elements {
		geoFence := []h3.GeoCoord{}
		h3Idxs := []h3.H3Index{}
		for _, latLng := range element.Points {
			h3Idx := h3.FromGeo(
				h3.GeoCoord{
					Latitude:  latLng.Lat,
					Longitude: latLng.Lng,
				},
				c.Resolution,
			)
			h3Idxs = append(h3Idxs, h3Idx)
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
			h3Idxs = append(h3Idxs, h3Idx)
		}

		for _, h3Idx := range h3Idxs {
			if _, ok := cellStatistics[h3Idx]; ok {
				for _, code := range element.FinalCodes {
					cellStatistics[h3Idx].POIs[code]++
				}
			}
		}
	}

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
