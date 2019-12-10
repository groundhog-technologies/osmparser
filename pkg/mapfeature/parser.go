package mapfeature

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"strings"
)

// Parser is parser for map features.
type Parser interface {
	Run() (MapFeatures, error)
}

// GetPrimartFeaturesParser return parser for PrimaryFeatures.
func GetPrimartFeaturesParser(url string) Parser {
	return &PrimaryFeaturesParser{
		URL: url,
	}
}

// PrimaryFeaturesParser .
type PrimaryFeaturesParser struct {
	URL string
}

func (p *PrimaryFeaturesParser) cleanStr(s string) string {
	space := regexp.MustCompile(`\s+`)
	newS := space.ReplaceAllString(s, "")
	newS = strings.TrimSuffix(newS, "\n")
	newS = strings.TrimSpace(newS)
	newS = strings.ToLower(newS)
	return newS
}

// Run .
func (p *PrimaryFeaturesParser) Run() (MapFeatures, error) {
	mapFeatures := MapFeatures{Key: "PrimaryFeatures", Values: make(map[string]MapFeatures)}
	// Get DOC.
	wikiURL := p.URL
	resp, err := http.Get(wikiURL)
	if err != nil {
		return mapFeatures, err
	}
	if resp.StatusCode != 200 {
		return mapFeatures, fmt.Errorf("resp.StatusCode != 200")
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return mapFeatures, err
	}

	// Parser.

	doc.Find("ul .tocsection-1").Each(func(i int, s *goquery.Selection) {
		s.Find("li.toclevel-2").Each(func(i int, s2 *goquery.Selection) {
			// Key
			key := p.cleanStr(s2.Find("a").First().Find("span.toctext").Text())
			mapFeatures.Values[key] = MapFeatures{
				Key:    key,
				Values: make(map[string]MapFeatures),
			}
			s2.Find("ul").Find("li.toclevel-3").Each(func(i int, s3 *goquery.Selection) {
				// subKey
				subKey := p.cleanStr(s3.Find("a").First().Find("span.toctext").Text())
				mapFeatures.Values[key].Values[subKey] = MapFeatures{
					Key:    subKey,
					Values: make(map[string]MapFeatures),
				}
			})
			mapFeatures.Values[key].Values["other"] = MapFeatures{
				Key:    "other",
				Values: make(map[string]MapFeatures),
			}
		})
	})

	doc.Find("table.wikitable").Each(func(i int, s *goquery.Selection) {
		subKey := ""
		s.Find("tbody").Find("tr").Each(func(i int, s2 *goquery.Selection) {
			if s2.Find("th").First().Text() == "key" {
				return
			}
			if s2.Find("th").Length() == 1 {
				subKey = p.cleanStr(s2.Find("th").First().Text())
				return
			}
			key := ""
			s2.Find("td").Each(func(i int, s3 *goquery.Selection) {
				if i == 0 {
					key = p.cleanStr(s3.Text())
				}
				if i == 1 {
					value := p.cleanStr(s3.Text())
					if _, ok := mapFeatures.Values[key]; ok {
						if _, ok := mapFeatures.Values[key].Values[subKey]; ok {
							mapFeatures.Values[key].Values[subKey].Values[value] = MapFeatures{
								Key: value,
							}
						} else {
							mapFeatures.Values[key].Values["other"].Values[value] = MapFeatures{
								Key: value,
							}
						}
					}
				}
			})
		})

	})
	return mapFeatures, nil
}
