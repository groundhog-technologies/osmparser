package osm

import (
	"github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
	"os"
	"osmparser/pkg/bitmask"
	"sync"
)

// PBFIndexParser .
type PBFIndexParser interface {
	PBFDataParser
	GetMap() *bitmask.PBFIndexMap
}

// NewPBFIndexer .
func NewPBFIndexer(pbfFile string, pbfIdxMap *bitmask.PBFIndexMap) PBFIndexParser {
	return &PBFIndexer{
		PBFFile:   pbfFile,
		PBFIdxMap: pbfIdxMap,
	}
}

// PBFIndexer .
type PBFIndexer struct {
	PBFFile   string
	PBFIdxMap *bitmask.PBFIndexMap
	MapLock   sync.RWMutex
}

// ReadNode .
func (p *PBFIndexer) ReadNode(n gosmparse.Node) {
	if len(n.Tags) > 0 {
		p.PBFIdxMap.Nodes.Insert(n.ID)
	}
}

// ReadWay .
func (p *PBFIndexer) ReadWay(w gosmparse.Way) {
	if len(w.Tags) > 0 {
		p.PBFIdxMap.Ways.Insert(w.ID)
		for _, nodeID := range w.NodeIDs {
			p.PBFIdxMap.WayRefs.Insert(nodeID)
		}
	}
}

// ReadRelation .
func (p *PBFIndexer) ReadRelation(r gosmparse.Relation) {
	if len(r.Tags) > 0 {
		var count = make(map[int]int64)
		for _, member := range r.Members {
			count[int(member.Type)]++
		}

		// Skip if relations contain 0 way.
		if count[1] == 0 {
			return
		}
		p.PBFIdxMap.Relations.Insert(r.ID)
		for _, member := range r.Members {
			switch member.Type {
			case 0:
				p.PBFIdxMap.RelNodes.Insert(member.ID)
			case 1:
				p.PBFIdxMap.RelWays.Insert(member.ID)
			case 2:
				p.PBFIdxMap.RelRelation.Insert(member.ID)
			}
		}
	}
}

// GetMap .
func (p *PBFIndexer) GetMap() *bitmask.PBFIndexMap {
	return p.PBFIdxMap
}

// Run .
func (p *PBFIndexer) Run() error {
	reader, err := os.Open(p.PBFFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	decoder := gosmparse.NewDecoder(reader)
	if err := decoder.Parse(p); err != nil {
		return err
	}
	return nil
}
