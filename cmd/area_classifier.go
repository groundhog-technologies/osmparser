package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uber/h3-go"
	"osm-parser/pkg/classifier"
	"osm-parser/pkg/entity"
	mapfeature "osm-parser/pkg/mapfeature"
	"osm-parser/pkg/osm"
	"sync"
	"time"
)

func init() {
	addOSMFlag(AreaClassifierCmd)
	AreaClassifierCmd.Flags().Int("h3_resolution", 9, "H3 resolution. See https://uber.github.io/h3/#/documentation/core-library/resolution-table .")
	AreaClassifierCmd.Flags().String("output_csv", "./classifier_output.csv", "output csv path.")
}

var taipeiGeoPolygon = h3.GeoPolygon{
	Geofence: []h3.GeoCoord{
		h3.GeoCoord{
			Latitude:  25.0747155493453,
			Longitude: 121.476860046387,
		},
		h3.GeoCoord{
			Latitude:  25.1058082610185,
			Longitude: 121.500205993652,
		},
		h3.GeoCoord{

			Latitude:  25.1014557570421,
			Longitude: 121.573677062988,
		},
		h3.GeoCoord{
			Latitude:  25.0467253568212,
			Longitude: 121.606636047363,
		},
		h3.GeoCoord{
			Latitude:  25.001927757185,
			Longitude: 121.59496307373,
		},
		h3.GeoCoord{
			Latitude:  24.9882363415471,
			Longitude: 121.562004089355,
		},
		h3.GeoCoord{
			Latitude:  24.9950822400259,
			Longitude: 121.49471282959,
		},
		h3.GeoCoord{
			Latitude:  25.0367717478887,
			Longitude: 121.460380554199,
		},
	},
}

// AreaClassifierCmd .
var AreaClassifierCmd = &cobra.Command{
	Use:   "area_classifier",
	Short: "Classifier area in earth.",
	RunE: func(cmd *cobra.Command, args []string) error {
		st := time.Now()

		// MapFeatures parser.
		url := viper.GetString("wiki_url")
		html := viper.GetString("wiki_html")
		parser := mapfeature.GetPrimartFeaturesParser(url, html)
		mapFeatures, err := parser.Run(false)
		if err != nil {
			return err
		}
		logrus.Infof("map features parser finish.")

		// OSM .pbf Data parser.
		handler, err := osm.NewDataHandler(mapFeatures)
		if err != nil {
			return err
		}

		wg := sync.WaitGroup{}
		elementChan := make(chan entity.Element)

		wg.Add(1)
		elements := []entity.Element{}
		go func() {
			defer wg.Done()
			for element := range elementChan {
				elements = append(elements, element)
			}
		}()

		if err := handler.Run(elementChan, viper.GetString("pbf_file")); err != nil {
			return err
		}
		wg.Wait()
		logrus.Infof("pbf parser finish.")

		// Classifier .
		c := classifier.AreaClassifier{
			Elements:    elements,
			Resolution:  viper.GetInt("h3_resolution"),
			MapFeatures: mapFeatures,
			GeoPolygon:  taipeiGeoPolygon,
			OutputCSV:   viper.GetString("output_csv"),
		}

		if err := c.Run(); err != nil {
			return err
		}
		logrus.Infof("Timeit: %.2f", time.Since(st).Seconds())
		return nil
	},
}
