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
	c.Provide(func() string {
		return "../../src/taiwan-latest.osm.pbf"
	})

	var m map[int64]int
	err := c.Invoke(func(parser PBFIndexParser) {
		parser.Run()
		m = parser.GetMap()
	})

	if err != nil {
		t.Error(err)
	}

	for k, v := range m {
		t.Log(k, v)
	}
}
