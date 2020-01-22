package osm

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/thomersch/gosmparse"
	"go.uber.org/dig"
	"math"
	"os"
	"osmparser/pkg/bitmask"
	"strconv"
	"sync"
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
		BatchSize:                params.BatchSize,
	}
}

// PBFParser .
type PBFParser struct {
	dig.In
	PBFFile  string
	PBFMasks *bitmask.PBFMasks
	// Indexer
	PBFIndexer               PBFDataParser
	PBFRelationMemberIndexer PBFDataParser
	// DB
	DB          *leveldb.DB
	LevelDBPath string
	Batch       *leveldb.Batch
	BatchSize   int

	// Chan
	ElementChan chan Element
}

// GetMap .
func (p *PBFParser) GetMap() *bitmask.PBFMasks {
	return p.PBFMasks
}

// Run .
func (p *PBFParser) Run() error {
	// Prepare
	db, err := leveldb.OpenFile(p.LevelDBPath, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	p.DB = db

	// Index .
	if err := p.PBFIndexer.Run(); err != nil {
		return err
	}
	if err := p.PBFRelationMemberIndexer.Run(); err != nil {
		return err
	}

	reader, err := os.Open(p.PBFFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Before run.
	p.Batch = new(leveldb.Batch)
	p.ElementChan = make(chan Element, 10)
	wg := sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		for element := range p.ElementChan {
			switch element.Type {
			case "Node":
				node := element.Node
				if p.PBFMasks.WayRefs.Has(node.ID) || p.PBFMasks.RelNodes.Has(node.ID) {
					id, val := p.nodeToBytes(node)
					// CacheQueue
					p.Batch.Put([]byte(id), []byte(val))

					if p.Batch.Len() > p.BatchSize {
						if err := p.cacheFlush(true); err != nil {
							logrus.Fatal(err)
						}
					}
				}
				if p.PBFMasks.Nodes.Has(node.ID) {
					fmt.Println(string(element.ToJSON()))
				}
			case "Way":
			case "Relation":
			}
		}
	}()

	decoder := gosmparse.NewDecoder(reader)
	if err := decoder.Parse(p); err != nil {
		return err
	}
	close(p.ElementChan)
	wg.Wait()
	return nil
}

// ReadNode .
func (p *PBFParser) ReadNode(n gosmparse.Node) {
	p.ElementChan <- Element{
		Type: "Node",
		Node: n,
	}
}

// ReadWay .
func (p *PBFParser) ReadWay(w gosmparse.Way) {
	p.ElementChan <- Element{
		Type: "Way",
		Way:  w,
	}
}

// ReadRelation .
func (p *PBFParser) ReadRelation(r gosmparse.Relation) {
	p.ElementChan <- Element{
		Type:     "Relation",
		Relation: r,
	}
}

func (p *PBFParser) nodeToBytes(n gosmparse.Node) (string, []byte) {
	var buf bytes.Buffer

	// Encoding lat as 64 bit float64 packed into 8 bytes.
	var latBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(latBytes, math.Float64bits(n.Lat))
	buf.Write(latBytes)

	// Encoding lng as 64 bit float64 packed into 8 bytes.
	var lonBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(latBytes, math.Float64bits(n.Lon))
	buf.Write(lonBytes)

	return strconv.FormatInt(n.ID, 10), buf.Bytes()
}

func (p *PBFParser) cacheFlush(sync bool) error {
	writeOpts := &opt.WriteOptions{
		NoWriteMerge: true,
		Sync:         sync,
	}
	err := p.DB.Write(p.Batch, writeOpts)
	if err != nil {
		return err
	}
	p.Batch.Reset()
	return nil
}
