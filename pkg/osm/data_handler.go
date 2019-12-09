package osm

import (
	"github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
	"os"
	// "strings"
	"sync"
)

// NewDataHandler return handler.
func NewDataHandler() (Handler, error) {
	return &DataHandler{
		NodeMap:      make(map[int64]gosmparse.Node),
		NodeMapMutex: sync.RWMutex{},
		ElementMap:   make(map[int64]Element),
	}, nil
}

// DataHandler .
// Implement the gosmparser.OSMReader interface here.
// Streaming data will call those functions.
type DataHandler struct {
	NodeMap      map[int64]gosmparse.Node
	NodeMapMutex sync.RWMutex
	ElementMap   map[int64]Element
}

// ReadNode .
func (d *DataHandler) ReadNode(n gosmparse.Node) {
	d.NodeMapMutex.Lock()
	if len(n.Element.Tags) > 0 {
		d.ElementMap[n.Element.ID] = Element{
			ID:   n.Element.ID,
			Type: "node",
			Points: []LatLng{
				LatLng{
					Lat: n.Lat,
					Lng: n.Lon,
				},
			},
			Tags: n.Element.Tags,
		}
	}
	d.NodeMapMutex.Unlock()

	d.NodeMapMutex.Lock()
	d.NodeMap[n.Element.ID] = n
	d.NodeMapMutex.Unlock()
}

// ReadWay .
func (d *DataHandler) ReadWay(w gosmparse.Way) {
	element := Element{
		ID:     w.Element.ID,
		Type:   "Way",
		Points: []LatLng{},
		Tags:   w.Element.Tags,
	}
	for _, nodeID := range w.NodeIDs {
		d.NodeMapMutex.Lock()
		if v, ok := d.NodeMap[nodeID]; ok {
			element.Points = append(
				element.Points,
				LatLng{Lat: v.Lat, Lng: v.Lon},
			)
			delete(d.NodeMap, nodeID)
		}
		d.NodeMapMutex.Unlock()
	}

	d.NodeMapMutex.Lock()
	d.ElementMap[w.Element.ID] = element
	d.NodeMapMutex.Unlock()
}

// ReadRelation .
func (d *DataHandler) ReadRelation(r gosmparse.Relation) {
}

// Run .
func (d *DataHandler) Run(pbfFile string) error {
	r, err := os.Open(pbfFile)
	if err != nil {
		return err
	}
	dec := gosmparse.NewDecoder(r)

	// Parse will block until it is done or an error occurs.
	err = dec.Parse(d)
	if err != nil {
		return err
	}
	d.NodeMap = nil

	logrus.Infof("%#v", dec)
	finalCodeCountMap := make(map[string]int)
	cellMap := make(map[string]Cell)
	for _, v := range d.ElementMap {
		if err := v.GenFinalCode(); err != nil {
			logrus.Error(err)
			continue
		}
		finalCodeCountMap[v.FinalCode]++
		cellMap[v.FinalCode] = v.Cell
	}
	for k, v := range finalCodeCountMap {
		logrus.Debugf("%v: %v : %v", k, v, cellMap[k])
	}
	logrus.Infof("%#v", len(d.ElementMap))
	logrus.Infof("%#v", len(finalCodeCountMap))
	return nil
}
