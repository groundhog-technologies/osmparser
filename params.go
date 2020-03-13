package osm

import (
	"github.com/groundhog-technologies/osmparser/pkg/bitmask"
	"github.com/groundhog-technologies/osmparser/pkg/element"
	"go.uber.org/dig"
)

// DefaultPBFParserParams .
type DefaultPBFParserParams struct {
	dig.In
	PBFFile  string            `name:"pbfFile"`
	PBFMasks *bitmask.PBFMasks `name:"pbfMasks"`
}

// PBFParserParams .
type PBFParserParams struct {
	dig.In
	LevelDBPath              string               `name:"levelDBPath"`
	PBFIndexer               PBFDataParser        `name:"pbfIndexer"`
	PBFRelationMemberIndexer PBFDataParser        `name:"pbfRelationMemberIndexer"`
	BatchSize                int                  `name:"batchSize"`
	OutputElementChan        chan element.Element `name:"outputElementChan"`
}
