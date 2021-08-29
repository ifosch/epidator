package epidactor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"github.com/ifosch/stationery/pkg/gdrive"
	"github.com/ifosch/stationery/pkg/stationery"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v2"
)

var Now = time.Now

var GetFeed = func(URL string) (string, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

var GetScript = func(episodeTag string) (string, error) {
	q := fmt.Sprintf("name contains '%v'", episodeTag)

	svc, err := gdrive.GetService(os.Getenv("DRIVE_CREDENTIALS_FILE"))
	if err != nil {
		return "", err
	}

	if len(q) == 0 {
		return "", fmt.Errorf("no matching scripts, please add a query returning one single document")
	}

	r, err := stationery.GetFiles(svc, q)
	if err != nil {
		return "", err
	}

	if len(r) > 1 {
		return "", fmt.Errorf("too many results. Query must return only one document, not %v", len(r))
	}

	content, err := stationery.ExportHTML(svc, r[0])
	if err != nil {
		return "", err
	}

	return content, nil
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
	FeedURL     string `yaml:"feedURL"`
	Cover       string `yaml:"cover"`
	Artist      string `yaml:"artist"`
	Album       string `yaml:"album"`
	MasterURL   string `yaml:"masterURL"`
	IntroURL    string `yaml:"introURL"`
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

func ExtractProperties(trackFileName, script, feed string, propertiesDefinitions *PropDefs) (properties map[string]interface{}, err error) {
	properties = map[string]interface{}{}

	scriptTree, err := htmlquery.Parse(strings.NewReader(script))
	if err != nil {
		return nil, err
	}
	feedTree, err := htmlquery.Parse(strings.NewReader(feed))
	if err != nil {
		return nil, err
	}

	for _, propDef := range propertiesDefinitions.Definitions {
		if propDef.List {
			htmlNodes := htmlquery.Find(scriptTree, propDef.Hook)
			contents := []string{}
			for _, htmlNode := range htmlNodes {
				contents = append(contents, htmlquery.InnerText(htmlNode))
			}
			properties[propDef.Name] = contents
		} else {
			htmlNode := htmlquery.FindOne(scriptTree, propDef.Hook)
			if propDef.Attribute != "" {
				properties[propDef.Name] = htmlquery.SelectAttr(htmlNode, propDef.Attribute)
			} else {
				properties[propDef.Name] = htmlquery.InnerText(htmlNode)
			}
		}
	}

	trackNo, err := ExtractTrackNo(feedTree, propertiesDefinitions)
	if err != nil {
		return nil, err
	}

	properties["trackNo"] = trackNo + 1
	properties["pubDate"] = Now()
	properties["cover"] = propertiesDefinitions.Cover
	properties["artist"] = propertiesDefinitions.Artist
	properties["album"] = propertiesDefinitions.Album
	properties["master"] = strings.Replace(propertiesDefinitions.MasterURL, "<FILE>", trackFileName, 1)
	properties["intro"] = propertiesDefinitions.IntroURL

	return properties, nil
}

func GetEpisodeDetails(trackName, propertiesDefinitionsYAML string) (map[string]interface{}, error) {
	propertiesDefinitions, err := NewPropDefs(propertiesDefinitionsYAML)
	if err != nil {
		return nil, err
	}

	script, err := GetScript(ExtractEpisodeTagFromTrack(trackName))
	if err != nil {
		return nil, err
	}

	feed, err := GetFeed(propertiesDefinitions.FeedURL)
	if err != nil {
		return nil, err
	}

	properties, err := ExtractProperties(trackName, script, feed, propertiesDefinitions)
	if err != nil {
		return nil, err
	}

	return properties, nil
}
