package epidactor

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v2"
)

type PropDefs struct {
	FeedURL          string `yaml:"feedURL"`
	MasterURLPattern string `yaml:"masterURLPattern"`
	DirectFields     struct {
		Cover    string `yaml:"cover"`
		Artist   string `yaml:"artist"`
		Album    string `yaml:"album"`
		IntroURL string `yaml:"introURL"`
	} `yaml:"directFields"`
	ScriptFieldHooks []struct {
		Name      string `yaml:"name"`
		Hook      string `yaml:"hook"`
		List      bool   `yaml:"list"`
		Attribute string `yaml:"attribute"`
	} `yaml:"scriptFieldHooks"`
	EpisodeScriptHooks map[string]string `yaml:"episodeScriptHooks"`
	trackName          string
	scriptTree         *html.Node
	feedTree           *html.Node
	properties         map[string]interface{}
}

func NewPropDefs(trackName, YAMLFile string) (*PropDefs, error) {
	f, err := ioutil.ReadFile(YAMLFile)
	if err != nil {
		return nil, err
	}

	propDefs := &PropDefs{
		trackName:  trackName,
		properties: map[string]interface{}{},
	}
	err = yaml.Unmarshal([]byte(f), propDefs)
	if err != nil {
		return nil, err
	}

	script, err := GetScript(propDefs.EpisodeScriptHook())
	if err != nil {
		return nil, err
	}

	propDefs.scriptTree, err = htmlquery.Parse(strings.NewReader(script))
	if err != nil {
		return nil, err
	}

	feed, err := GetFeed(propDefs.FeedURL)
	if err != nil {
		return nil, err
	}

	propDefs.feedTree, err = htmlquery.Parse(strings.NewReader(feed))
	if err != nil {
		return nil, err
	}

	err = propDefs.ExtractProperties()
	if err != nil {
		return nil, err
	}

	return propDefs, nil
}

func (pd *PropDefs) EpisodeScriptHook() string {
	numberRE := regexp.MustCompile("[0-9]+")

	tag := ""
	for k, v := range pd.EpisodeScriptHooks {
		if strings.Contains(pd.trackName, k) {
			tag = v
		}
	}
	if tag == "" {
		tag = pd.EpisodeScriptHooks["default"]
	}

	return fmt.Sprintf("%s %s", tag, numberRE.FindString(pd.trackName))
}

func (pd *PropDefs) TrackNo() (int, error) {
	expr, err := xpath.Compile("count(//item)")
	if err != nil {
		return 0, err
	}

	trackNo := int(expr.Evaluate(htmlquery.CreateXPathNavigator(pd.feedTree)).(float64)) + 1

	return trackNo, nil
}

func (pd *PropDefs) ExtractPropertiesFromScript() {
	for _, hook := range pd.ScriptFieldHooks {
		if hook.List {
			htmlNodes := htmlquery.Find(pd.scriptTree, hook.Hook)
			contents := []string{}
			for _, htmlNode := range htmlNodes {
				contents = append(contents, htmlquery.InnerText(htmlNode))
			}
			pd.properties[hook.Name] = contents
		} else {
			htmlNode := htmlquery.FindOne(pd.scriptTree, hook.Hook)
			if hook.Attribute != "" {
				pd.properties[hook.Name] = htmlquery.SelectAttr(htmlNode, hook.Attribute)
			} else {
				pd.properties[hook.Name] = htmlquery.InnerText(htmlNode)
			}
		}
	}
}

func (pd *PropDefs) ExtractPropertiesFromFeed() error {
	trackNo, err := pd.TrackNo()
	if err != nil {
		return err
	}

	pd.properties["trackNo"] = trackNo
	return nil
}

func (pd *PropDefs) ExtractDirectProperties() {
	pd.properties["cover"] = pd.DirectFields.Cover
	pd.properties["artist"] = pd.DirectFields.Artist
	pd.properties["album"] = pd.DirectFields.Album
	pd.properties["intro"] = pd.DirectFields.IntroURL
}

func (pd *PropDefs) ExtractProperties() (err error) {
	pd.ExtractPropertiesFromScript()
	pd.ExtractDirectProperties()
	pd.properties["pubDate"] = Now()
	pd.properties["master"] = strings.Replace(pd.MasterURLPattern, "<FILE>", pd.trackName, 1)
	err = pd.ExtractPropertiesFromFeed()

	return err
}
