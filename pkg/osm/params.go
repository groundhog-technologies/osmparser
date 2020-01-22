package osm

import (
	"go.uber.org/dig"
	"osmparser/pkg/bitmask"
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
	LevelDBPath              string        `name:"levelDBPath"`
	PBFIndexer               PBFDataParser `name:"pbfIndexer"`
	PBFRelationMemberIndexer PBFDataParser `name:"pbfRelationMemberIndexer"`
}
