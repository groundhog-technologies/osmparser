package osm

import (
	"github.com/groundhog-technologies/osmparser/pkg/bitmask"
	"go.uber.org/dig"
	"testing"
)

func TestPBFRelationMemberIndexer(t *testing.T) {
	c := dig.New()
	c.Provide(
		func() string {
			return "./src/testing.pbf"
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

	err := c.Invoke(func(parser PBFDataParser) {
		if err := parser.Run(); err != nil {
			t.Error(err)
		}
	})

	if err != nil {
		t.Error(err)
	}
}
