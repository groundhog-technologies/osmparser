package osm

import (
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"osmparser/pkg/bitmask"
	"testing"
)

func init() {
	logrus.AddHook(filename.NewHook())
}

func TestPBFParser(t *testing.T) {

	c := dig.New()

	// Default params .
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
	// Params
	c.Provide(NewPBFIndexer, dig.Name("pbfIndexer"))
	c.Provide(NewPBFRelationMemberIndexer, dig.Name("pbfRelationMemberIndexer"))
	c.Provide(
		func() string {
			return "/tmp/osmparser"
		},
		dig.Name("levelDBPath"),
	)
	c.Provide(
		func() int {
			return 5000
		},
		dig.Name("batchSize"),
	)

	c.Provide(NewPBFParser)

	err := c.Invoke(func(parser PBFDataParser) {
		if err := parser.Run(); err != nil {
			t.Error(err)
		}
	})

	if err != nil {
		t.Error(err)
	}
}
