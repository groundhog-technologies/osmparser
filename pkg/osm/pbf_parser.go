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
	var finishNode, finishWay bool

	// Sync
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		numNode := 0
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
				if p.PBFMasks.Nodes.Has(element.Node.ID) {
					// fmt.Println(string(element.ToJSON()))
				}
				numNode++
				if numNode%1000000 == 0 {
					logrus.Infof("Num: %v", numNode)
				}
			case "Way":
				// Flush outstanding node batches before processing any ways.
				if !finishNode {
					finishNode = true
					logrus.Info("Finish Node")
					if p.Batch.Len() > 1 {
						p.cacheFlush(true)
					}
				}

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

				if p.PBFMasks.Ways.Has(element.Way.ID) {
					elements, err := p.cacheLookupWayElements(&element.Way)
					// skip ways which fail to denormalize.
					if err != nil {
						continue
					}
					element.Elements = elements
				}
			case "Relation":
				if !finishWay {
					finishWay = true
					logrus.Info("Finish Way")
					if p.Batch.Len() > 1 {
						p.cacheFlush(true)
					}
				}

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

				if p.PBFMasks.Relations.Has(element.Relation.ID) {
					elements, err := p.cacheLookupRelationElements(&element.Relation)
					if err != nil {
						logrus.Error(err)
						continue
					}
					element.Elements = elements
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

func (p *PBFParser) checkBatch() {
	if p.Batch.Len() > p.BatchSize {
		if err := p.cacheFlush(true); err != nil {
			logrus.Fatal(err)
		}
	}
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

func (p *PBFParser) bytesToNodeElement(data []byte) Element {

	node := gosmparse.Node{}
	// bytes to LatLon .
	var latBytes = append([]byte{}, data[0:8]...)
	var lat = math.Float64frombits(binary.BigEndian.Uint64(latBytes))
	node.Lat = lat

	var lonBytes = append([]byte{}, data[8:16]...)
	var lon = math.Float64frombits(binary.BigEndian.Uint64(lonBytes))
	node.Lon = lon
	return Element{
		Type: "Node",
		Node: node,
	}
}

func (p *PBFParser) cacheLookupWayElements(way *gosmparse.Way) ([]Element, error) {
	var elements []Element
	for _, nodeID := range way.NodeIDs {
		strID := strconv.FormatInt(nodeID, 10)
		data, err := p.DB.Get([]byte(strID), nil)
		if err != nil {
			return []Element{}, err
		}
		element := p.bytesToNodeElement(data)
		elements = append(elements, element)

	}
	return elements, nil
}

func (p *PBFParser) cacheLookupRelationElements(relation *gosmparse.Relation) ([]Element, error) {
	var elements []Element
	for _, member := range relation.Members {
		strID := strconv.FormatInt(member.ID, 10)
		switch member.Type {
		case 0: // Node
			nodeBytes, err := p.DB.Get([]byte(strID), nil)
			if err != nil {
				logrus.Error(err)
				return []Element{}, err
			}
			element := p.bytesToNodeElement(nodeBytes)
			elements = append(elements, element)
		case 1: // Way
			elementByte, err := p.DB.Get([]byte("W"+strID), nil)
			if err != nil {
				logrus.Error(err)
				return []Element{}, err
			}
			element, err := ByteToElement(elementByte)
			if err != nil {
				logrus.Error(err)
				return []Element{}, err
			}
			nodeElements, err := p.cacheLookupWayElements(&element.Way)
			if err != nil {
				logrus.Error(err)
				return []Element{}, err
			}
			element.Elements = nodeElements
			element.Role = member.Role
			elements = append(elements, element)
		case 2: // Relation
			elementByte, err := p.DB.Get([]byte("R"+strID), nil)
			if err != nil {
				logrus.Error(err)
				return []Element{}, err
			}
			element, err := ByteToElement(elementByte)
			if err != nil {
				logrus.Error(err)
				return []Element{}, err
			}
			newElements, err := p.cacheLookupRelationElements(&element.Relation)
			if err != nil {
				logrus.Error(err)
				return []Element{}, err
			}
			element.Elements = newElements
			element.Role = member.Role
			elements = append(elements, element)
		}
	}
	return elements, nil
}
