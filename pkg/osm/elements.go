package osm

import (
	"fmt"
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

// GenInfo .
func (e *Element) GenInfo(mapFeatures mapfeature.MapFeatures) error {
	for k, v := range e.Tags {
		if l1Features, ok := mapFeatures.Values[strings.ToLower(k)]; ok {
			var codeL1, codeL2, codeL3 string
			codeL1 = k

			for _, l2Features := range l1Features.Values {
				if l3Features, ok := l2Features.Values[v]; ok {
					codeL2 = l2Features.Key
					codeL3 = l3Features.Key
				}
			}
			if codeL2 == "" {
				codeL2 = "other"
			}
			if codeL3 == "" {
				codeL3 = "other"
			}
			finalCode := fmt.Sprintf("%v:%v:%v", codeL1, codeL2, codeL3)
			e.FinalCodes = append(e.FinalCodes, finalCode)
		}
	}
	return nil
}
