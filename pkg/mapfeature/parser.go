package mapfeature

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// MapFeaturesParser is parser for map features.
type MapFeaturesParser interface {
	Run(useURL bool) (MapFeatures, error)
}

// GetPrimartFeaturesParser return parser for PrimaryFeatures.
func GetPrimartFeaturesParser(url string, html string) MapFeaturesParser {
	return &PrimaryFeaturesParser{
		URL:  url,
		HTML: html,
	}
}

// PrimaryFeaturesParser .
type PrimaryFeaturesParser struct {
	URL  string
	HTML string
}

// cleanHTMLID turn layer key to unite format.
func (p *PrimaryFeaturesParser) cleanHTMLID(s string) string {
	newS := strings.TrimSuffix(s, "\n")
	newS = strings.TrimSpace(newS)
	space := regexp.MustCompile(`\s+`)
	newS = space.ReplaceAllString(newS, "_")
	return newS
}

// cleanKey clean mapFeatures key value to unite format.
func (p *PrimaryFeaturesParser) cleanKey(s string) string {
	newS := strings.TrimSuffix(s, "\n")
	newS = strings.TrimSpace(newS)
	// Remove space.
	spaceGexp := regexp.MustCompile(`\s+`)
	newS = spaceGexp.ReplaceAllString(newS, "_")
	// replace all special Characters to _ .
	specialCharactersGexp := regexp.MustCompile(`\W+`)
	newS = specialCharactersGexp.ReplaceAllString(newS, "_")

	// Replace duplicate _ to one.
	underscoreDuplicateGexp := regexp.MustCompile(`\_+`)
	newS = underscoreDuplicateGexp.ReplaceAllString(newS, "_")

	newS = strings.ToLower(newS)
	newS = strings.ReplaceAll(newS, ",_", "_")
	return newS
}

// Run .
func (p *PrimaryFeaturesParser) Run(useURL bool) (MapFeatures, error) {
	mapFeatures := MapFeatures{Key: "PrimaryFeatures", Values: make(map[string]MapFeatures)}
	// Get DOC.
	var docReader io.Reader
	if useURL {
		resp, err := http.Get(p.URL)
		if err != nil {
			return mapFeatures, err
		}
		if resp.StatusCode != 200 {
			return mapFeatures, fmt.Errorf("resp.StatusCode != 200")
		}
		defer resp.Body.Close()
		docReader = resp.Body
		logrus.Infof("Get map features from %v", p.URL)
	} else {
		f, err := os.Open(p.HTML)
		if err != nil {
			return mapFeatures, err
		}
		defer f.Close()
		docReader = f
		logrus.Infof("Get map features from %v", p.HTML)
	}

	doc, err := goquery.NewDocumentFromReader(docReader)
	if err != nil {
		return mapFeatures, err
	}

	// Parser for Layer 1 & Layer 2.
	doc.Find("ul .tocsection-1").Each(func(i int, s *goquery.Selection) {
		s.Find("li.toclevel-2").Each(func(i int, s2 *goquery.Selection) {
			// Layer1 key.
			layer1Key := p.cleanHTMLID(s2.Find("a").First().Find("span.toctext").Text())
			mapFeatures.Values[layer1Key] = MapFeatures{
				Key:    layer1Key,
				Values: make(map[string]MapFeatures),
			}
			s2.Find("ul").Find("li.toclevel-3").Each(func(i int, s3 *goquery.Selection) {
				// Layer2 key.
				layer2Key := p.cleanHTMLID(s3.Find("a").First().Find("span.toctext").Text())
				mapFeatures.Values[layer1Key].Values[layer2Key] = MapFeatures{
					Key:    layer2Key,
					Values: make(map[string]MapFeatures),
				}
			})

			// Add "other" for every layer1.
			mapFeatures.Values[layer1Key].Values["other"] = MapFeatures{
				Key:    "other",
				Values: make(map[string]MapFeatures),
			}
		})
	})

	// Parser for Layer 3.
	for layer1Key := range mapFeatures.Values {
		docL1 := doc.Find(fmt.Sprintf("h3 span#%v", layer1Key))
		if docL1.Length() > 0 {
			// Case 1.
			if goquery.NodeName(docL1.Parent().Next().Next()) == "table" {
				table := docL1.Parent().Next().Next()
				layer2Key := ""
				table.Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
					if s.Find("th").First().Text() == "key" {
						return
					}
					if s.Find("th").Find("H4").Length() == 1 {
						layer2Key = p.cleanHTMLID(s.Find("th").Find("h4").First().Text())
						return
					}
					s.Find("td").Each(func(i int, s2 *goquery.Selection) {
						if i == 1 {
							layer3Key := p.cleanHTMLID(s2.Text())
							if _, ok := mapFeatures.Values[layer1Key].Values[layer2Key]; ok {
								mapFeatures.Values[layer1Key].Values[layer2Key].Values[layer3Key] = MapFeatures{
									Key: layer3Key,
								}
							} else {
								mapFeatures.Values[layer1Key].Values["other"].Values[layer3Key] = MapFeatures{
									Key: layer3Key,
								}
							}
						}
					})
				})
			}
			// Case 2.
			if goquery.NodeName(docL1.Parent().Next().Next()) == "h4" {
				docL1.Parent().NextAll().EachWithBreak(func(i int, s *goquery.Selection) bool {
					if goquery.NodeName(s) == "h3" {
						return false
					}
					if goquery.NodeName(s) == "h4" {
						if tags, ok := s.Next().Attr("data-taginfo-taglist-tags"); ok {
							for _, tag := range strings.Split(strings.Split(tags, "=")[1], ",") {
								layer2Key := p.cleanHTMLID(s.Text())
								layer3Key := p.cleanHTMLID(tag)
								if _, ok := mapFeatures.Values[layer1Key].Values[layer2Key]; ok {
									mapFeatures.Values[layer1Key].Values[layer2Key].Values[layer3Key] = MapFeatures{
										Key: layer3Key,
									}
								} else {
									mapFeatures.Values[layer1Key].Values["other"].Values[layer3Key] = MapFeatures{
										Key: layer3Key,
									}
								}
							}
						}
					}
					return true
				})
			}
			// Case 3.
			if docL1.Parent().Next().Next().HasClass("taglist") {
				newDoc := docL1.Parent().Next().Next()
				if tags, ok := newDoc.Attr("data-taginfo-taglist-tags"); ok {
					for _, tag := range strings.Split(strings.Split(tags, "=")[1], ",") {
						layer3Key := p.cleanHTMLID(tag)
						mapFeatures.Values[layer1Key].Values["other"].Values[layer3Key] = MapFeatures{
							Key: layer3Key,
						}
					}
				}
			}
		}
	}

	// Clean key.
	newMapFeatures := MapFeatures{Key: "PrimaryFeatures", Values: make(map[string]MapFeatures)}
	for k, v := range mapFeatures.Values {
		cleanKey := p.cleanKey(k)
		newMapFeatures.Values[cleanKey] = MapFeatures{Key: cleanKey, Values: make(map[string]MapFeatures)}
		for k2, v2 := range v.Values {
			cleanKey2 := p.cleanKey(k2)
			newMapFeatures.Values[cleanKey].Values[cleanKey2] = MapFeatures{Key: cleanKey2, Values: make(map[string]MapFeatures)}
			for k3 := range v2.Values {
				cleanKey3 := p.cleanKey(k3)
				newMapFeatures.Values[cleanKey].Values[cleanKey2].Values[cleanKey3] = MapFeatures{Key: cleanKey3}
			}
		}
	}
	return newMapFeatures, nil
}
