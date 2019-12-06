package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"osm-parser/pkg/osm"
)

func init() {
	OSMParserCmd.Flags().String("pbf_file", "./src/taiwan-latest.osm.pbf", "pbf file path.")
}

// OSMParserCmd .
var OSMParserCmd = &cobra.Command{
	Use:   "osm_parser",
	Short: "Parse osm data.",
	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := osm.NewDataHandler()
		if err != nil {
			return err
		}
		if err := handler.Run(viper.GetString("pbf_file")); err != nil {
			return err
		}
		return nil
	},
}
