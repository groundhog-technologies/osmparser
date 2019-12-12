package osm

import (
	"github.com/thomersch/gosmparse"
)

// Handler is interface for handler osm data.
type Handler interface {
	gosmparse.OSMReader
	Run(dataChan chan Element, pbfFile string) error
}
