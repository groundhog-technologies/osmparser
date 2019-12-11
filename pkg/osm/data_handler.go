package osm

import (
	"github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
	"os"
	"osm-parser/pkg/mapfeature"
	"sync"
	// "strings"
)

// NewDataHandler return handler.
func NewDataHandler(mapFeatures mapfeature.MapFeatures) (Handler, error) {
	return &DataHandler{
		NodeMap:      make(map[int64]gosmparse.Node),
		NodeMapMutex: sync.RWMutex{},
		ElementMap:   make(map[int64]Element),
		MapFeatures:  mapFeatures,
		ElementChan:  make(chan Element),
		NodeChan:     make(chan gosmparse.Node),
		WayChan:      make(chan gosmparse.Way),
	}, nil
}

// DataHandler .
// Implement the gosmparser.OSMReader interface here.
// Streaming data will call those functions.
type DataHandler struct {
	NodeMap      map[int64]gosmparse.Node
	NodeMapMutex sync.RWMutex
	ElementMap   map[int64]Element
	MapFeatures  mapfeature.MapFeatures
	ElementChan  chan Element
	NodeChan     chan gosmparse.Node
	WayChan      chan gosmparse.Way
}

// ReadNode .
func (d *DataHandler) ReadNode(n gosmparse.Node) {
	d.NodeChan <- n
}

// ReadWay .
func (d *DataHandler) ReadWay(w gosmparse.Way) {
	d.WayChan <- w
}

// ReadRelation .
func (d *DataHandler) ReadRelation(r gosmparse.Relation) {
}

// Run .
func (d *DataHandler) Run(pbfFile string) error {
	wg := sync.WaitGroup{}

	// Element worker.
	wg.Add(1)
	go func() {
		defer wg.Done()
		eMap := make(map[string]int)
		for element := range d.ElementChan {
			if err := element.GenInfo(d.MapFeatures); err != nil {
				logrus.Error(err)
				continue
			}
			logrus.Debug(element.Tags, "  ", element.Points)
			logrus.Info(element.FinalCodes)
			for _, code := range element.FinalCodes {
				eMap[code]++
			}
		}
		for k, v := range eMap {
			logrus.Debugf("%v: %v", k, v)
		}
		logrus.Info(len(eMap))
	}()

	// Node worker.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for n := range d.NodeChan {
			if len(n.Element.Tags) > 0 {
				element := Element{
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
				d.ElementChan <- element
			}
			d.NodeMapMutex.Lock()
			d.NodeMap[n.Element.ID] = n
			d.NodeMapMutex.Unlock()
		}
	}()

	// Way worker.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for w := range d.WayChan {
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
			d.ElementChan <- element
		}
	}()

	r, err := os.Open(pbfFile)
	if err != nil {
		return err
	}
	dec := gosmparse.NewDecoder(r)
	// Parse will block until it is done or an error occurs.
	err = dec.Parse(d)

	close(d.NodeChan)
	close(d.WayChan)
	close(d.ElementChan)
	wg.Wait()
	if err != nil {
		return err
	}

	return nil
}
