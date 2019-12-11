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

func (p *PrimaryFeaturesParser) cleanHTMLID(s string) string {
	newS := strings.TrimSuffix(s, "\n")
	newS = strings.TrimSpace(newS)
	space := regexp.MustCompile(`\s+`)
	newS = space.ReplaceAllString(newS, "_")
	return newS
}

func (p *PrimaryFeaturesParser) cleanKey(s string) string {
	newS := strings.TrimSuffix(s, "\n")
	newS = strings.TrimSpace(newS)
	space := regexp.MustCompile(`\s+`)
	newS = space.ReplaceAllString(newS, "_")
	newS = strings.ToLower(newS)
	newS = strings.ReplaceAll(newS, ",_", "_")
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
			key := p.cleanHTMLID(s2.Find("a").First().Find("span.toctext").Text())
			mapFeatures.Values[key] = MapFeatures{
				Key:    key,
				Values: make(map[string]MapFeatures),
			}
			s2.Find("ul").Find("li.toclevel-3").Each(func(i int, s3 *goquery.Selection) {
				// subKey
				subKey := p.cleanHTMLID(s3.Find("a").First().Find("span.toctext").Text())
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

	for k := range mapFeatures.Values {
		docL1 := doc.Find(fmt.Sprintf("h3 span#%v", k))
		if docL1.Length() > 0 {
			if goquery.NodeName(docL1.Parent().Next().Next()) == "table" {
				table := docL1.Parent().Next().Next()
				subKey := ""
				table.Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
					if s.Find("th").First().Text() == "key" {
						return
					}
					if s.Find("th").Length() == 1 {
						subKey = p.cleanHTMLID(s.Find("th").First().Text())
						return
					}
					s.Find("td").Each(func(i int, s2 *goquery.Selection) {
						if i == 1 {
							value := p.cleanHTMLID(s2.Text())
							if _, ok := mapFeatures.Values[k].Values[subKey]; ok {
								mapFeatures.Values[k].Values[subKey].Values[value] = MapFeatures{
									Key: value,
								}
							} else {
								mapFeatures.Values[k].Values["other"].Values[value] = MapFeatures{
									Key: value,
								}
							}
						}
					})
				})
			}
			if goquery.NodeName(docL1.Parent().Next().Next()) == "h4" {
				docL1.Parent().NextAll().EachWithBreak(func(i int, s *goquery.Selection) bool {
					if goquery.NodeName(s) == "h3" {
						return false
					}
					if goquery.NodeName(s) == "h4" {
						if tags, ok := s.Next().Attr("data-taginfo-taglist-tags"); ok {
							for _, tag := range strings.Split(strings.Split(tags, "=")[1], ",") {
								subKey := p.cleanHTMLID(s.Text())
								value := p.cleanHTMLID(tag)
								if _, ok := mapFeatures.Values[k].Values[subKey]; ok {
									mapFeatures.Values[k].Values[subKey].Values[value] = MapFeatures{
										Key: value,
									}
								} else {
									mapFeatures.Values[k].Values["other"].Values[value] = MapFeatures{
										Key: value,
									}
								}
							}
						}
					}
					return true
				})
			}
			if docL1.Parent().Next().Next().HasClass("taglist") {
				newDoc := docL1.Parent().Next().Next()
				if tags, ok := newDoc.Attr("data-taginfo-taglist-tags"); ok {
					for _, tag := range strings.Split(strings.Split(tags, "=")[1], ",") {
						value := p.cleanHTMLID(tag)
						mapFeatures.Values[k].Values["other"].Values[value] = MapFeatures{
							Key: value,
						}
					}
				}
			}
		}
	}

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
