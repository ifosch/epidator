package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/ifosch/stationery/pkg/gdrive"
	"github.com/ifosch/stationery/pkg/stationery"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v2"
)

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

func ExtractProperties(doc *html.Node, propertiesDefinitions *PropDefs) (properties map[string]interface{}) {
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

	return
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

	doc, err := GetScript(ExtractEpisodeTagFromTrack(os.Args[1]))
	if err != nil {
		log.Fatal(err)
	}

	propertiesDefinitions, err := NewPropDefs("propertiesDefinitions.yaml")
	if err != nil {
		log.Fatal(err)
	}

	PrintYAML(ExtractProperties(doc, propertiesDefinitions))
}
