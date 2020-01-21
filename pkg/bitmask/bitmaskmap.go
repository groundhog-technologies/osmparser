package bitmask

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
)

// PBFIndexMap - struct to hold common masks .
type PBFIndexMap struct {
	Nodes       *Bitmask
	Ways        *Bitmask
	Relations   *Bitmask
	WayRefs     *Bitmask
	RelNodes    *Bitmask
	RelWays     *Bitmask
	RelRelation *Bitmask
}

// NewPBFIndexMap - constructor
func NewPBFIndexMap() *PBFIndexMap {
	return &PBFIndexMap{
		Nodes:       NewBitMask(),
		Ways:        NewBitMask(),
		Relations:   NewBitMask(),
		WayRefs:     NewBitMask(),
		RelNodes:    NewBitMask(),
		RelWays:     NewBitMask(),
		RelRelation: NewBitMask(),
	}
}

// WriteTo - write to destination
func (m *PBFIndexMap) WriteTo(sink io.Writer) (int64, error) {
	encoder := gob.NewEncoder(sink)
	err := encoder.Encode(m)
	return 0, err
}

// ReadFrom - read from destination
func (m *PBFIndexMap) ReadFrom(tap io.Reader) (int64, error) {
	decoder := gob.NewDecoder(tap)
	err := decoder.Decode(m)
	return 0, err
}

// WriteToFile - write to disk
func (m *PBFIndexMap) WriteToFile(path string) {
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	m.WriteTo(file)
	log.Println("wrote bitmask:", path)
}

// ReadFromFile - read from disk
func (m *PBFIndexMap) ReadFromFile(path string) {

	// bitmask file doesn't exist
	if _, err := os.Stat(path); err != nil {
		fmt.Println("bitmask file not found:", path)
		os.Exit(1)
	}

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	m.ReadFrom(file)
	log.Println("read bitmask:", path)
}

// Print -- print debug stats
func (m PBFIndexMap) Print() {
	k := reflect.TypeOf(m)
	v := reflect.ValueOf(m)
	for i := 0; i < k.NumField(); i++ {
		key := k.Field(i).Name
		val := v.Field(i).Interface()
		fmt.Printf("%s: %v\n", key, (val.(*Bitmask)).Len())
	}
}
