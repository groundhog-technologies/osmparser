package osm

import (
	"fmt"
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
}

// ReadNode .
func (d *DataHandler) ReadNode(n gosmparse.Node) {
	if len(n.Element.Tags) > 0 {
		fmt.Printf("Node: %#v\n", n)
	}
	d.NodeMapMutex.Lock()
	d.NodeMap[n.Element.ID] = n
	d.NodeMapMutex.Unlock()
}

// ReadWay .
func (d *DataHandler) ReadWay(w gosmparse.Way) {
	fmt.Printf("Way: %#v\n", w)
	d.NodeMapMutex.Lock()
	for _, nodeID := range w.NodeIDs {
		fmt.Println(nodeID, d.NodeMap[nodeID])
	}
	d.NodeMapMutex.Unlock()
}

// ReadRelation .
func (d *DataHandler) ReadRelation(r gosmparse.Relation) {
	// fmt.Printf("Relation: %#v\n", r)
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
	return nil
}
