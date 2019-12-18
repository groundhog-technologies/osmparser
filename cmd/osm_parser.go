package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"osm-parser/pkg/entity"
	"osm-parser/pkg/mapfeature"
	"osm-parser/pkg/osm"
	"sync"
)

func init() {
	addOSMFlag(OSMParserCmd)
}

// OSMParserCmd .
var OSMParserCmd = &cobra.Command{
	Use:   "osm_parser",
	Short: "Parse osm data.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get MapFeatures.
		url := viper.GetString("wiki_url")
		html := viper.GetString("wiki_html")
		parser := mapfeature.GetPrimartFeaturesParser(url, html)
		mapFeatures, err := parser.Run(false)
		if err != nil {
			return err
		}

		// Data parser .
		handler, err := osm.NewDataHandler(mapFeatures)
		if err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		elementChan := make(chan entity.Element)

		wg.Add(1)
		go func() {
			defer wg.Done()
			idx := 0
			for element := range elementChan {
				idx++
				logrus.Debug(idx, " ", element.Tags, "  ", element.FinalCodes)
			}
		}()

		if err := handler.Run(elementChan, viper.GetString("pbf_file")); err != nil {
			return err
		}

		wg.Wait()
		return nil
	},
}
