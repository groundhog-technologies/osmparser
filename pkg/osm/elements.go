package osm

import (
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"hash/fnv"
	"osm-parser/pkg/mapfeature"
	"strings"
)

// LatLng .
type LatLng struct {
	Lat float64
	Lng float64
}

// Element .
type Element struct {
	ID         int64
	Type       string
	Points     []LatLng
	Tags       map[string]string
	FinalCodes []string
}

// GenFinalCode .
// Turn tags to md5Hash.
func (e *Element) GenFinalCode(mapFeatures mapfeature.MapFeatures) error {

	for k, v := range e.Tags {
		if subV, ok := mapFeatures.Values[strings.ToLower(k)]; ok {
			var tagKey, tagValue string
			tagKey = k

			for level2Tag, level2TagV := range subV.Values {
				if _, ok := level2TagV.Values[v]; ok {
					tagValue = level2Tag
				}
			}
			if tagValue == "" {
				tagValue = "other"
			}
			hasher := fnv.New32a()
			hasher.Write([]byte(fmt.Sprintf("%v:%v", tagKey, tagValue)))
			finalCode := hex.EncodeToString(hasher.Sum(nil))
			e.FinalCodes = append(e.FinalCodes, finalCode)
		}
	}
	logrus.Debugf("%+v %v", e.Tags, e.FinalCodes)
	return nil
}
