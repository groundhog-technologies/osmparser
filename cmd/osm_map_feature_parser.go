package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	mapfeature "osm-parser/pkg/mapfeature"
)

func init() {
	OSMMapFeatureParserCmd.Flags().String("wiki_url", "https://wiki.openstreetmap.org/wiki/Map_Features", "Map feature wiki url.")

}

// OSMMapFeatureParserCmd .
var OSMMapFeatureParserCmd = &cobra.Command{
	Use:   "osm_map_feature_parser",
	Short: "Parse osm map feature data.",
	RunE: func(cmd *cobra.Command, args []string) error {
		url := viper.GetString("wiki_url")
		html := viper.GetString("wiki_html")
		parser := mapfeature.GetPrimartFeaturesParser(url, html)
		mapFeatures, err := parser.Run(false)
		if err != nil {
			return err
		}

		for k, v := range mapFeatures.Values {
			logrus.Debug("----------")
			for k2, v2 := range v.Values {
				for k3 := range v2.Values {
					logrus.Debugf("key: %v subKey: %v value: %v", k, k2, k3)
				}
			}
		}
		return nil
	},
}
