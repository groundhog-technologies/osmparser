package osm

import (
	"github.com/thomersch/gosmparse"
)

// Handler is interface for handler osm data.
type Handler interface {
	gosmparse.OSMReader
	Run(pbfFile string) error
}
