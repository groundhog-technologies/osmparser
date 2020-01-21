package osm

import (
	"github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
	// "os"
	"osmparser/pkg/bitmask"
)

// NewPBFParser .
func NewPBFParser(pbfFile string, pbfIndexer PBFIndexParser, pbfRelationMemberIndexer PBFRelationMemberIndexParser, pbfMasks *bitmask.PBFMasks) PBFDataParser {
	return &PBFParser{
		PBFFile:                  pbfFile,
		PBFIndexer:               pbfIndexer,
		PBFRelationMemberIndexer: pbfRelationMemberIndexer,
		PBFMasks:                 pbfMasks,
	}
}

// PBFParser .
type PBFParser struct {
	PBFFile                  string
	PBFIndexer               PBFIndexParser
	PBFRelationMemberIndexer PBFIndexParser
	PBFMasks                 *bitmask.PBFMasks
}

// GetMap .
func (p *PBFParser) GetMap() *bitmask.PBFMasks {
	return p.PBFMasks
}

// Run .
func (p *PBFParser) Run() error {
	logrus.Infof("%#v", p)
	p.PBFMasks.Print()
	if err := p.PBFIndexer.Run(); err != nil {
		return err
	}
	p.PBFMasks.Print()
	if err := p.PBFRelationMemberIndexer.Run(); err != nil {
		return err
	}
	p.PBFMasks.Print()
	return nil
}

// ReadNode .
func (p *PBFParser) ReadNode(n gosmparse.Node) {}

// ReadWay .
func (p *PBFParser) ReadWay(w gosmparse.Way) {}

// ReadRelation .
func (p *PBFParser) ReadRelation(r gosmparse.Relation) {}
