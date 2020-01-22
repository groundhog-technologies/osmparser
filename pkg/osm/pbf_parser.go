package osm

import (
	"github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
	"go.uber.org/dig"
	"os"
	"osmparser/pkg/bitmask"
)

// NewPBFParser .
func NewPBFParser(
	defaultParams DefaultPBFParserParams,
	params PBFParserParams,
) PBFDataParser {
	return &PBFParser{
		PBFFile:                  defaultParams.PBFFile,
		PBFMasks:                 defaultParams.PBFMasks,
		PBFIndexer:               params.PBFIndexer,
		LevelDBPath:              params.LevelDBPath,
		PBFRelationMemberIndexer: params.PBFRelationMemberIndexer,
	}
}

// PBFParser .
type PBFParser struct {
	dig.In
	PBFFile                  string
	LevelDBPath              string
	PBFIndexer               PBFDataParser
	PBFRelationMemberIndexer PBFDataParser
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
func (p *PBFParser) ReadNode(n gosmparse.Node) {
	if p.PBFMasks.WayRefs.Has(n.ID) || p.PBFMasks.RelNodes.Has(n.ID) {
	}
}

// ReadWay .
func (p *PBFParser) ReadWay(w gosmparse.Way) {}

// ReadRelation .
func (p *PBFParser) ReadRelation(r gosmparse.Relation) {}
