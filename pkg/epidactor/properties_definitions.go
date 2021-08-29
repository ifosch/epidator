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
	EpisodeHooks map[string]string `yaml:"episodeHooks"`
	trackName    string
	scriptTree   *html.Node
	feedTree     *html.Node
}

func NewPropDefs(trackName, YAMLFile string) (*PropDefs, error) {
	f, err := ioutil.ReadFile(YAMLFile)
	if err != nil {
		return nil, err
	}

	propDefs := &PropDefs{
		trackName: trackName,
	}
	err = yaml.Unmarshal([]byte(f), propDefs)
	if err != nil {
		return nil, err
	}

	script, err := GetScript(propDefs.EpisodeHook())
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

	return propDefs, nil
}

func (pd *PropDefs) EpisodeHook() string {
	numberRE := regexp.MustCompile("[0-9]+")

	tag := ""
	for k, v := range pd.EpisodeHooks {
		if strings.Contains(pd.trackName, k) {
			tag = v
		}
	}
	if tag == "" {
		tag = pd.EpisodeHooks["default"]
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
