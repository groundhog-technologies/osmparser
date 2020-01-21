package osm

import (
	"github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
	"os"
	"osm-parser/pkg/bitmask"
	"sync"
)

// PBFIndexParser .
type PBFIndexParser interface {
	PBFDataParser
	GetMap() bitmask.PBFIndexParser
}

// NewPBFIndexer .
func NewPBFIndexer(pbfFile string) PBFIndexParser {
	return &PBFIndexer{
		PBFFile: pbfFile,
		Map:     make(map[int64]int),
		MapLock: sync.RWMutex{},
	}
}

// PBFIndexer .
type PBFIndexer struct {
	PBFFile string
	Map     map[int64]int
	MapLock sync.RWMutex
}

// Insert .
func (p *PBFIndexer) Insert(index int64) {
	p.MapLock.Lock()
	defer p.MapLock.Unlock()
	p.Map[index] = 1
}

// ReadNode .
func (p *PBFIndexer) ReadNode(n gosmparse.Node) {
	p.Insert(n.ID)
}

// ReadWay .
func (p *PBFIndexer) ReadWay(w gosmparse.Way) {
	// logrus.Info(w)
	p.Insert(w.ID)
}

// ReadRelation .
func (p *PBFIndexer) ReadRelation(r gosmparse.Relation) {
	// logrus.Info(r)
}

// GetMap .
func (p *PBFIndexer) GetMap() map[int64]int {
	return p.Map
}

// Run .
func (p *PBFIndexer) Run() error {
	logrus.Info(p.PBFFile)
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
