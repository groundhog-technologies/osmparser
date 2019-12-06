ZONE=asia/taiwan


##@ Download

get-osm-pbf:  ## Download lasest version osm pbf file use ZONE variable. ex, make get-osm-pbf ZONE=asia/taiwan
	wget http://download.geofabrik.de/${ZONE}-latest.osm.pbf -P ./src

##@ Help

.PHONY: help

help:  ## Display this help
		@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
