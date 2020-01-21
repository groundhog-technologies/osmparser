package osm

import (
	"go.uber.org/dig"
	"testing"
)

func TestPBFIndexer(t *testing.T) {
	c := dig.New()
	c.Provide(NewPBFIndexer)
	c.Provide(func() string {
		return "../../src/taiwan-latest.osm.pbf"
	})

	err := c.Invoke(func(parser PBFDataParser) {
		parser.Run()
	})

	if err != nil {
		t.Error(err)
	}
}
