package osm

import (
	"github.com/thomersch/gosmparse"
	"osm-parser/pkg/entity"
)

// Handler is interface for handler osm data.
type Handler interface {
	gosmparse.OSMReader
	Run(dataChan chan entity.Element, pbfFile string) error
}
