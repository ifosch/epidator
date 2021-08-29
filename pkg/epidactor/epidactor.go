package epidactor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"github.com/ifosch/stationery/pkg/gdrive"
	"github.com/ifosch/stationery/pkg/stationery"
	"golang.org/x/net/html"
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

func ExtractTrackNo(feed *html.Node, propertiesDefinitions *PropDefs) (int, error) {
	expr, err := xpath.Compile("count(//item)")
	if err != nil {
		return 0, err
	}

	trackNo := int(expr.Evaluate(htmlquery.CreateXPathNavigator(feed)).(float64))

	return trackNo, nil
}

func ExtractProperties(propertiesDefinitions *PropDefs) (properties map[string]interface{}, err error) {
	properties = map[string]interface{}{}

	for _, propDef := range propertiesDefinitions.Definitions {
		if propDef.List {
			htmlNodes := htmlquery.Find(propertiesDefinitions.scriptTree, propDef.Hook)
			contents := []string{}
			for _, htmlNode := range htmlNodes {
				contents = append(contents, htmlquery.InnerText(htmlNode))
			}
			properties[propDef.Name] = contents
		} else {
			htmlNode := htmlquery.FindOne(propertiesDefinitions.scriptTree, propDef.Hook)
			if propDef.Attribute != "" {
				properties[propDef.Name] = htmlquery.SelectAttr(htmlNode, propDef.Attribute)
			} else {
				properties[propDef.Name] = htmlquery.InnerText(htmlNode)
			}
		}
	}

	trackNo, err := ExtractTrackNo(propertiesDefinitions.feedTree, propertiesDefinitions)
	if err != nil {
		return nil, err
	}

	properties["trackNo"] = trackNo + 1
	properties["pubDate"] = Now()
	properties["cover"] = propertiesDefinitions.Cover
	properties["artist"] = propertiesDefinitions.Artist
	properties["album"] = propertiesDefinitions.Album
	properties["master"] = strings.Replace(propertiesDefinitions.MasterURL, "<FILE>", propertiesDefinitions.trackName, 1)
	properties["intro"] = propertiesDefinitions.IntroURL

	return properties, nil
}

func GetEpisodeDetails(trackName, propertiesDefinitionsYAML string) (map[string]interface{}, error) {
	propertiesDefinitions, err := NewPropDefs(trackName, propertiesDefinitionsYAML)
	if err != nil {
		return nil, err
	}

	properties, err := ExtractProperties(propertiesDefinitions)
	if err != nil {
		return nil, err
	}

	return properties, nil
}
