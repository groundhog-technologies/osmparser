package osm

import (
	"github.com/groundhog-technologies/osmparser/pkg/bitmask"
	"github.com/thomersch/gosmparse"
	"os"
	"sync"
)

// NewPBFRelationMemberIndexer .
func NewPBFRelationMemberIndexer(params DefaultPBFParserParams) PBFDataParser {
	return &PBFRelationMemberIndexer{
		PBFFile:  params.PBFFile,
		PBFMasks: params.PBFMasks,
	}
}

// PBFRelationMemberIndexer .
type PBFRelationMemberIndexer struct {
	PBFFile  string
	PBFMasks *bitmask.PBFMasks
	MapLock  sync.RWMutex
}

// Run .
func (p *PBFRelationMemberIndexer) Run() error {
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

// ReadNode .
func (p *PBFRelationMemberIndexer) ReadNode(n gosmparse.Node) {}

// ReadWay .
func (p *PBFRelationMemberIndexer) ReadWay(w gosmparse.Way) {
	if p.PBFMasks.RelWays.Has(w.ID) {
		for _, nodeID := range w.NodeIDs {
			p.PBFMasks.RelNodes.Insert(nodeID)
		}
	}
}

// ReadRelation .
func (p *PBFRelationMemberIndexer) ReadRelation(r gosmparse.Relation) {}
