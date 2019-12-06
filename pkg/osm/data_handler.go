package osm

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
	"os"
	"sync"
)

// NewDataHandler return handler.
func NewDataHandler() (Handler, error) {
	return &DataHandler{
		NodeMap:      make(map[int64]gosmparse.Node),
		NodeMapMutex: sync.RWMutex{},
	}, nil
}

// DataHandler .
// Implement the gosmparser.OSMReader interface here.
// Streaming data will call those functions.
type DataHandler struct {
	NodeMap      map[int64]gosmparse.Node
	NodeMapMutex sync.RWMutex
	Elements     []Element
}

// ReadNode .
func (d *DataHandler) ReadNode(n gosmparse.Node) {
	d.NodeMapMutex.Lock()
	if len(n.Element.Tags) > 0 {
		d.Elements = append(
			d.Elements,
			Element{
				ID:   n.Element.ID,
				Type: "node",
				Points: []LatLng{
					LatLng{
						Lat: n.Lat,
						Lng: n.Lon,
					},
				},
				Tags: n.Element.Tags,
			},
		)
	}
	d.NodeMap[n.Element.ID] = n
	d.NodeMapMutex.Unlock()
}

// ReadWay .
func (d *DataHandler) ReadWay(w gosmparse.Way) {
	// fmt.Printf("Way: %#v\n", w)
	d.NodeMapMutex.Lock()
	for _, nodeID := range w.NodeIDs {
		if v, ok := d.NodeMap[nodeID]; ok {
			// logrus.Debug(nodeID, v)
			if len(v.Element.Tags) > 0 {
				logrus.Debugf("WayTags: %v\n NodeTags: %v", w.Tags, v.Element.Tags)
			}
		}
	}
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
	fmt.Printf("%#v", dec)
	fmt.Printf("%#v", len(d.Elements))
	return nil
}
