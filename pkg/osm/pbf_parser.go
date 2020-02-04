package osm

import (
	"bytes"
	"encoding/binary"
	// "fmt"
	"github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/thomersch/gosmparse"
	"go.uber.org/dig"
	"io"
	"math"
	"os"
	"osmparser/pkg/bitmask"
	"osmparser/pkg/element"
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
		OutputElementChan:        params.OutputElementChan,
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
	ElementChan       chan element.Element
	OutputElementChan chan element.Element
}

// GetMap .
func (p *PBFParser) GetMap() *bitmask.PBFMasks {
	return p.PBFMasks
}

// Run .
func (p *PBFParser) Run() error {
	// Prepare
	db, err := leveldb.OpenFile(
		p.LevelDBPath,
		// Disable cache to avoid memory issue.
		&opt.Options{DisableBlockCache: true},
	)
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
	logrus.Info("Finish index")

	reader, err := os.Open(p.PBFFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	// First round.
	// Put way refs, relation member in to db.

	// Before run.
	p.Batch = new(leveldb.Batch)
	p.ElementChan = make(chan element.Element, 10)

	// Sync
	firstRoundWg := sync.WaitGroup{}
	firstRoundWg.Add(1)

	go func() {
		defer firstRoundWg.Done()
		for element := range p.ElementChan {
			switch element.Type {
			case "Node":
				// Write way refs and relation member nodes to db.
				if p.PBFMasks.WayRefs.Has(element.Node.ID) || p.PBFMasks.RelNodes.Has(element.Node.ID) {
					id, val := p.nodeToBytes(element.Node)
					// CacheQueue
					p.Batch.Put(
						[]byte(id),
						[]byte(val),
					)
					p.checkBatch()
				}
			case "Way":
				// Write relation member way to db.
				if p.PBFMasks.RelWays.Has(element.Way.ID) {
					elementByte, err := element.ToByte()
					if err != nil {
						logrus.Error(err)
						continue
					}
					p.Batch.Put(
						[]byte("W"+strconv.FormatInt(element.Way.ID, 10)),
						elementByte,
					)
					p.checkBatch()
				}
			case "Relation":
				// Write relation Member into db.
				if p.PBFMasks.RelRelation.Has(element.Relation.ID) {
					elementByte, err := element.ToByte()
					if err != nil {
						logrus.Error(err)
						continue
					}
					p.Batch.Put(
						[]byte("R"+strconv.FormatInt(element.Relation.ID, 10)),
						elementByte,
					)
					p.checkBatch()
				}
			}
		}
	}()
	firstRoundDecoder := gosmparse.NewDecoder(reader)
	if err := firstRoundDecoder.Parse(p); err != nil {
		return err
	}
	close(p.ElementChan)
	firstRoundWg.Wait()
	p.cacheFlush(true)
	logrus.Info("Finish first round.")
	reader.Seek(io.SeekStart, 0) // rewind file.

	// Final round.
	// Real process for parse pbf file.

	// Before run.
	p.Batch = new(leveldb.Batch)
	p.ElementChan = make(chan element.Element, 10)

	// Sync
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(p.OutputElementChan)
		for emt := range p.ElementChan {
			switch emt.Type {
			case "Node":
				if p.PBFMasks.Nodes.Has(emt.Node.ID) {
					p.OutputElementChan <- emt
					// fmt.Println(string(emt.ToJSON()))
				}
			case "Way":
				if p.PBFMasks.Ways.Has(emt.Way.ID) {
					emts, err := p.cacheLookupWayElements(&emt.Way)
					// skip ways which fail to denormalize.
					if err != nil {
						continue
					}
					emt.Elements = emts
					p.OutputElementChan <- emt
				}
			case "Relation":
				if p.PBFMasks.Relations.Has(emt.Relation.ID) {
					emts, err := p.cacheLookupRelationElements(&emt.Relation, []int64{})
					// skip ways which fail to denormalize.
					if err != nil {
						continue
					}
					emt.Elements = emts
					p.OutputElementChan <- emt
				}
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
	p.ElementChan <- element.Element{
		Type: "Node",
		Node: n,
	}
}

// ReadWay .
func (p *PBFParser) ReadWay(w gosmparse.Way) {
	p.ElementChan <- element.Element{
		Type: "Way",
		Way:  w,
	}
}

// ReadRelation .
func (p *PBFParser) ReadRelation(r gosmparse.Relation) {
	p.ElementChan <- element.Element{
		Type:     "Relation",
		Relation: r,
	}
}

// checkBatch check if need flush batch.
func (p *PBFParser) checkBatch() {
	if p.Batch.Len() > p.BatchSize {
		if err := p.cacheFlush(true); err != nil {
			logrus.Fatal(err)
		}
	}
}

// cacheFlush flush batch write to db.
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

// nodeToBytes transfrom node to bytes.
func (p *PBFParser) nodeToBytes(n gosmparse.Node) (string, []byte) {
	var buf bytes.Buffer

	// Encoding lat as 64 bit float64 packed into 8 bytes.
	var latBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(latBytes, math.Float64bits(n.Lat))
	buf.Write(latBytes)

	// Encoding lng as 64 bit float64 packed into 8 bytes.
	var lonBytes = make([]byte, 8)
	binary.BigEndian.PutUint64(lonBytes, math.Float64bits(n.Lon))
	buf.Write(lonBytes)

	return strconv.FormatInt(n.ID, 10), buf.Bytes()
}

// bytesToNodeElement transfrom node from bytes to element.
func (p *PBFParser) bytesToNodeElement(data []byte) element.Element {

	node := gosmparse.Node{}
	// bytes to LatLon .
	var latBytes = append([]byte{}, data[0:8]...)
	var lat = math.Float64frombits(binary.BigEndian.Uint64(latBytes))
	node.Lat = lat

	var lonBytes = append([]byte{}, data[8:16]...)
	var lon = math.Float64frombits(binary.BigEndian.Uint64(lonBytes))
	node.Lon = lon
	return element.Element{
		Type: "Node",
		Node: node,
	}
}

// cacheLookupWayElements get refs node from db.
func (p *PBFParser) cacheLookupWayElements(way *gosmparse.Way) ([]element.Element, error) {
	var emts []element.Element
	for _, nodeID := range way.NodeIDs {
		strID := strconv.FormatInt(nodeID, 10)
		data, err := p.DB.Get(
			[]byte(strID),
			nil,
		)
		if err != nil {
			return []element.Element{}, err
		}
		e := p.bytesToNodeElement(data)
		emts = append(emts, e)
	}
	return emts, nil
}

func (p *PBFParser) cacheLookupRelationElements(relation *gosmparse.Relation, blacklist []int64) ([]element.Element, error) {
	var emts []element.Element

	// blacklist to avoid recursive relation member.
	blacklist = append(blacklist, relation.ID)
	for _, member := range relation.Members {
		strID := strconv.FormatInt(member.ID, 10)
		switch member.Type {
		case 0: // Node
			nodeBytes, err := p.DB.Get(
				[]byte(strID),
				nil,
			)
			if err != nil {
				logrus.Error(err)
				return []element.Element{}, err
			}
			emt := p.bytesToNodeElement(nodeBytes)
			emts = append(emts, emt)
		case 1: // Way
			// Get element from db.
			elementByte, err := p.DB.Get(
				[]byte("W"+strID),
				nil,
			)
			if err != nil {
				logrus.Error(err)
				return []element.Element{}, err
			}
			emt, err := element.ByteToElement(elementByte)
			if err != nil {
				logrus.Error(err)
				return []element.Element{}, err
			}
			// Get ref nodes from db.
			nodeElements, err := p.cacheLookupWayElements(&emt.Way)
			if err != nil {
				logrus.Error(err)
				return []element.Element{}, err
			}
			emt.Elements = nodeElements
			emt.Role = member.Role
			emts = append(emts, emt)
		case 2: // Relation
			// Skip if relation recursive. A -> B -> A
			var recursive bool
			for _, blackID := range blacklist {
				if member.ID == blackID {
					recursive = true
				}
			}
			if recursive {
				continue
			}

			// Get element from db.
			elementByte, err := p.DB.Get(
				[]byte("R"+strID),
				nil,
			)
			if err != nil {
				logrus.Error(err)
				return []element.Element{}, err
			}
			emt, err := element.ByteToElement(elementByte)
			if err != nil {
				logrus.Error(err)
				return []element.Element{}, err
			}

			// Get relation member emts.
			newElements, err := p.cacheLookupRelationElements(&emt.Relation, blacklist)
			if err != nil {
				logrus.Error(err)
				return []element.Element{}, err
			}
			emt.Elements = newElements
			emt.Role = member.Role
			emts = append(emts, emt)
		}
	}
	return emts, nil
}
