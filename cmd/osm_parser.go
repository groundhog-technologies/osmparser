package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"osm-parser/pkg/mapfeature"
	"osm-parser/pkg/osm"
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
		parser := mapfeature.GetPrimartFeaturesParser(url)
		mapFeatures, err := parser.Run()
		if err != nil {
			return err
		}

		// Data parser .
		handler, err := osm.NewDataHandler(mapFeatures)
		if err != nil {
			return err
		}
		if err := handler.Run(viper.GetString("pbf_file")); err != nil {
			return err
		}
		return nil
	},
}
