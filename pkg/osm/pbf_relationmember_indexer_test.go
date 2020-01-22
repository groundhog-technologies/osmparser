package osm

import (
	"go.uber.org/dig"
	"osmparser/pkg/bitmask"
	"testing"
)

func TestPBFRelationMemberIndexer(t *testing.T) {
	c := dig.New()
	c.Provide(
		func() string {
			return "../../src/taiwan-latest.osm.pbf"
		},
		dig.Name("pbfFile"),
	)
	c.Provide(
		func() *bitmask.PBFMasks {
			return bitmask.NewPBFMasks()
		},
		dig.Name("pbfMasks"),
	)
	c.Provide(NewPBFRelationMemberIndexer)

	err := c.Invoke(func(parser PBFIndexParser) {
		if err := parser.Run(); err != nil {
			t.Error(err)
		}
	})

	if err != nil {
		t.Error(err)
	}
}
