package osm

import (
	"github.com/thomersch/gosmparse"
)

// PBFDataParser .
type PBFDataParser interface {
	gosmparse.OSMReader
	Run() error
}
