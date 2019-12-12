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
func (d *DataHandler) Run(dataChan chan Element, pbfFile string) error {
	wg := sync.WaitGroup{}

	doneChan := make(chan int)
	defer close(doneChan)

	// Element worker.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for element := range d.ElementChan {
			if err := element.GenInfo(d.MapFeatures); err != nil {
				logrus.Error(err)
				continue
			}
			dataChan <- element
		}
		close(dataChan)
		logrus.Info("Close dataChan")
		doneChan <- 1
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
		doneChan <- 1
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
		doneChan <- 1
	}()

	r, err := os.Open(pbfFile)
	if err != nil {
		return err
	}
	dec := gosmparse.NewDecoder(r)
	// Parse will block until it is done or an error occurs.
	err = dec.Parse(d)
	close(d.NodeChan)
	logrus.Info("Close NodeChan")

	finishWorker := 0
	finish := false
	for {
		select {
		case <-doneChan:
			finishWorker++
			switch finishWorker {
			case 1:
				close(d.WayChan)
				// Release.
				d.NodeMap = nil
				logrus.Info("Close WayChan")
			case 2:
				close(d.ElementChan)
				logrus.Info("Close ElementChan")
			case 3:
				finish = true
			}
		}
		if finish {
			break
		}
	}

	wg.Wait()
	if err != nil {
		return err
	}
	return nil
}
