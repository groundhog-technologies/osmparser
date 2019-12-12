package cmd

import (
	"github.com/spf13/cobra"
)

func addOSMFlag(cmd *cobra.Command) {
	cmd.Flags().String("wiki_url", "https://wiki.openstreetmap.org/wiki/Map_Features", "Map feature wiki url.")
	cmd.Flags().String("pbf_file", "./src/taiwan-latest.osm.pbf", "pbf file path.")
}
