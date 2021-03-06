package osm

import (
	"github.com/groundhog-technologies/osmparser/pkg/bitmask"
	"github.com/groundhog-technologies/osmparser/pkg/element"
	"github.com/onrik/logrus/filename"
	"github.com/sirupsen/logrus"
	"go.uber.org/dig"
	"sync"
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
	outputElementChan := make(chan element.Element)
	c.Provide(
		func() chan element.Element { return outputElementChan },
		dig.Name("outputElementChan"),
	)

	err := c.Invoke(func(parser PBFDataParser) error {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			num := 0
			for _ = range outputElementChan {
				num++
				if num%100000 == 0 {
					logrus.Infof("Element: %v", num)
				}
			}
		}()
		if err := parser.Run(); err != nil {
			return err
		}
		wg.Wait()
		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
