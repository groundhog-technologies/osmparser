GO ?= go
TESTFOLDER := $(shell $(GO) list ./... | grep -E 'osmparser$$' | grep -v examples)
TESTTAGS ?= ""
ZONE=asia/taiwan

##@ test

.PHONY: test
test:  ## Run test
	echo "mode: count" > coverage.out
	for d in $(TESTFOLDER); do \
		$(GO) test -tags $(TESTTAGS) -v -covermode=count -coverprofile=profile.out $$d > tmp.out; \
		cat tmp.out; \
		if grep -q "^--- FAIL" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "build failed" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "setup failed" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		fi; \
		if [ -f profile.out ]; then \
			cat profile.out | grep -v "mode:" >> coverage.out; \
			rm profile.out; \
		fi; \
	done


##@ Download

.PHONY: download

get-osm-pbf:  ## Download lasest version osm pbf file use ZONE variable. ex, make get-osm-pbf ZONE=asia/taiwan
	wget http://download.geofabrik.de/${ZONE}-latest.osm.pbf -P ./src

get-testing-pbf:  ## Download testing pbf file for testing.
	wget -nc http://download.geofabrik.de/asia/maldives-latest.osm.pbf -O ./src/testing.pbf

##@ Help

.PHONY: help

help:  ## Display this help
		@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
