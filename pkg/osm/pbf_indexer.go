package osm

import (
	"github.com/thomersch/gosmparse"
	"os"
	"osmparser/pkg/bitmask"
	"sync"
)

// PBFIndexParser .
type PBFIndexParser interface {
	PBFDataParser
	GetMap() *bitmask.PBFMasks
}

// NewPBFIndexer .
func NewPBFIndexer(pbfFile string, pbfMasks *bitmask.PBFMasks) PBFIndexParser {
	return &PBFIndexer{
		PBFFile:  pbfFile,
		PBFMasks: pbfMasks,
	}
}

// PBFIndexer .
type PBFIndexer struct {
	PBFFile  string
	PBFMasks *bitmask.PBFMasks
	MapLock  sync.RWMutex
}

// ReadNode .
func (p *PBFIndexer) ReadNode(n gosmparse.Node) {
	if len(n.Tags) > 0 {
		p.PBFMasks.Nodes.Insert(n.ID)
	}
}

// ReadWay .
func (p *PBFIndexer) ReadWay(w gosmparse.Way) {
	if len(w.Tags) > 0 {
		p.PBFMasks.Ways.Insert(w.ID)
		for _, nodeID := range w.NodeIDs {
			p.PBFMasks.WayRefs.Insert(nodeID)
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
		p.PBFMasks.Relations.Insert(r.ID)
		for _, member := range r.Members {
			switch member.Type {
			case 0:
				p.PBFMasks.RelNodes.Insert(member.ID)
			case 1:
				p.PBFMasks.RelWays.Insert(member.ID)
			case 2:
				p.PBFMasks.RelRelation.Insert(member.ID)
			}
		}
	}
}

// GetMap .
func (p *PBFIndexer) GetMap() *bitmask.PBFMasks {
	return p.PBFMasks
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
