package osm

import (
	"github.com/thomersch/gosmparse"
	"osmparser/pkg/bitmask"
)

// PBFDataParser .
type PBFDataParser interface {
	gosmparse.OSMReader
	Run() error
}

// PBFIndexParser .
type PBFIndexParser interface {
	PBFDataParser
	GetMap() *bitmask.PBFMasks
}
