package osm

import (
	"github.com/sirupsen/logrus"
	"github.com/thomersch/gosmparse"
	"os"
)

// NewPBFIndexer .
func NewPBFIndexer(pbfFile string) PBFDataParser {
	return &PBFIndexer{
		PBFFile: pbfFile,
	}
}

// PBFIndexer .
type PBFIndexer struct {
	PBFFile string
}

// ReadNode .
func (p *PBFIndexer) ReadNode(n gosmparse.Node) {
	logrus.Info(n)
}

// ReadWay .
func (p *PBFIndexer) ReadWay(w gosmparse.Way) {
	// logrus.Info(w)
}

// ReadRelation .
func (p *PBFIndexer) ReadRelation(r gosmparse.Relation) {
	// logrus.Info(r)
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
