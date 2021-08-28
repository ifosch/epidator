package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"github.com/ifosch/stationery/pkg/gdrive"
	"github.com/ifosch/stationery/pkg/stationery"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v2"
)

func GetFeed(URL string) (*html.Node, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func GetScript(episodeTag string) (*html.Node, error) {
	q := fmt.Sprintf("name contains '%v'", episodeTag)

	svc, err := gdrive.GetService(os.Getenv("DRIVE_CREDENTIALS_FILE"))
	if err != nil {
		return nil, err
	}

	if len(q) == 0 {
		return nil, fmt.Errorf("no matching scripts, please add a query returning one single document")
	}

	r, err := stationery.GetFiles(svc, q)
	if err != nil {
		return nil, err
	}

	if len(r) > 1 {
		return nil, fmt.Errorf("too many results. Query must return only one document, not %v", len(r))
	}

	content, err := stationery.ExportHTML(svc, r[0])
	if err != nil {
		return nil, err
	}

	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func ExtractEpisodeTagFromTrack(trackName string) (episodeTag string) {
	numberRE := regexp.MustCompile("[0-9]+")

	tag := "Podcast"
	if strings.Contains(trackName, "pildora") {
		tag = "Píldora"
	} else if strings.Contains(trackName, "colaboracion") {
		tag = "Colaboración"
	}

	return fmt.Sprintf("%s %s", tag, numberRE.FindString(trackName))
}

type PropDefs struct {
	URL         string `yaml:"feedURL"`
	Definitions []struct {
		Name      string `yaml:"name"`
		Hook      string `yaml:"hook"`
		List      bool   `yaml:"list"`
		Attribute string `yaml:"attribute"`
	} `yaml:"propDefs"`
}

func NewPropDefs(YAMLFile string) (*PropDefs, error) {
	f, err := ioutil.ReadFile(YAMLFile)
	if err != nil {
		return nil, err
	}

	propDefs := &PropDefs{}
	err = yaml.Unmarshal([]byte(f), propDefs)
	if err != nil {
		return nil, err
	}

	return propDefs, nil
}

func ExtractTrackNo(feed *html.Node, propertiesDefinitions *PropDefs) (int, error) {
	expr, err := xpath.Compile("count(//item)")
	if err != nil {
		return 0, err
	}

	trackNo := int(expr.Evaluate(htmlquery.CreateXPathNavigator(feed)).(float64))

	return trackNo, nil
}

func ExtractProperties(doc, feed *html.Node, propertiesDefinitions *PropDefs) (properties map[string]interface{}, err error) {
	properties = map[string]interface{}{}

	for _, propDef := range propertiesDefinitions.Definitions {
		if propDef.List {
			htmlNodes := htmlquery.Find(doc, propDef.Hook)
			contents := []string{}
			for _, htmlNode := range htmlNodes {
				contents = append(contents, htmlquery.InnerText(htmlNode))
			}
			properties[propDef.Name] = contents
		} else {
			htmlNode := htmlquery.FindOne(doc, propDef.Hook)
			if propDef.Attribute != "" {
				properties[propDef.Name] = htmlquery.SelectAttr(htmlNode, propDef.Attribute)
			} else {
				properties[propDef.Name] = htmlquery.InnerText(htmlNode)
			}
		}
	}

	trackNo, err := ExtractTrackNo(feed, propertiesDefinitions)
	if err != nil {
		return nil, err
	}

	properties["trackNo"] = trackNo + 1

	return properties, nil
}

func PrintYAML(properties map[string]interface{}) error {
	outputData, err := yaml.Marshal(properties)
	if err != nil {
		return err
	}

	fmt.Println(string(outputData))
	return nil
}

func main() {
	log.SetFlags(0)

	propertiesDefinitions, err := NewPropDefs("propertiesDefinitions.yaml")
	if err != nil {
		log.Fatal(err)
	}

	script, err := GetScript(ExtractEpisodeTagFromTrack(os.Args[1]))
	if err != nil {
		log.Fatal(err)
	}

	feed, err := GetFeed(propertiesDefinitions.URL)
	if err != nil {
		log.Fatal(err)
	}

	properties, err := ExtractProperties(script, feed, propertiesDefinitions)
	if err != nil {
		log.Fatal(err)
	}

	PrintYAML(properties)
}
