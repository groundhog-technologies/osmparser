package osm

import (
	"go.uber.org/dig"
	"osmparser/pkg/bitmask"
	"testing"
)

func TestPBFIndexer(t *testing.T) {
	c := dig.New()
	c.Provide(NewPBFIndexer)
	c.Provide(func() string {
		return "../../src/taiwan-latest.osm.pbf"
	})
	c.Provide(bitmask.NewPBFMasks)

	var m *bitmask.PBFMasks
	err := c.Invoke(func(parser PBFIndexParser) {
		if err := parser.Run(); err != nil {
			t.Error(err)
		}
		m = parser.GetMap()
	})

	if err != nil {
		t.Error(err)
	}
	t.Log(m)
}
