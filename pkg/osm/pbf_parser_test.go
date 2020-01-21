package osm

import (
	"go.uber.org/dig"
	"osmparser/pkg/bitmask"
	"testing"
)

func TestPBFParser(t *testing.T) {

	c := dig.New()
	c.Provide(NewPBFIndexer)
	c.Provide(NewPBFRelationMemberIndexer)
	c.Provide(NewPBFParser)
	c.Provide(func() string {
		return "../../src/taiwan-latest.osm.pbf"
	})
	masks := bitmask.NewPBFMasks()
	c.Provide(func() *bitmask.PBFMasks {
		return masks
	})

	err := c.Invoke(func(parser PBFDataParser) {
		if err := parser.Run(); err != nil {
			t.Error(err)
		}
	})

	if err != nil {
		t.Error(err)
	}
}
