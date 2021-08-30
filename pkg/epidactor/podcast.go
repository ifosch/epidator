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

type Podcast struct {
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
	details            map[string]interface{}
}

func NewPodcast(trackName, YAMLFile string) (*Podcast, error) {
	f, err := ioutil.ReadFile(YAMLFile)
	if err != nil {
		return nil, err
	}

	podcast := &Podcast{
		trackName: trackName,
		details:   map[string]interface{}{},
	}
	err = yaml.Unmarshal([]byte(f), podcast)
	if err != nil {
		return nil, err
	}

	script, err := GetScript(podcast.EpisodeScriptHook())
	if err != nil {
		return nil, err
	}

	podcast.scriptTree, err = htmlquery.Parse(strings.NewReader(script))
	if err != nil {
		return nil, err
	}

	feed, err := GetFeed(podcast.FeedURL)
	if err != nil {
		return nil, err
	}

	podcast.feedTree, err = htmlquery.Parse(strings.NewReader(feed))
	if err != nil {
		return nil, err
	}

	err = podcast.ExtractProperties()
	if err != nil {
		return nil, err
	}

	return podcast, nil
}

func (p *Podcast) EpisodeScriptHook() string {
	numberRE := regexp.MustCompile("[0-9]+")

	tag := ""
	for k, v := range p.EpisodeScriptHooks {
		if strings.Contains(p.trackName, k) {
			tag = v
		}
	}
	if tag == "" {
		tag = p.EpisodeScriptHooks["default"]
	}

	return fmt.Sprintf("%s %s", tag, numberRE.FindString(p.trackName))
}

func (p *Podcast) TrackNo() (int, error) {
	expr, err := xpath.Compile("count(//item)")
	if err != nil {
		return 0, err
	}

	trackNo := int(expr.Evaluate(htmlquery.CreateXPathNavigator(p.feedTree)).(float64)) + 1

	return trackNo, nil
}

func (p *Podcast) ExtractPropertiesFromScript() {
	for _, hook := range p.ScriptFieldHooks {
		if hook.List {
			htmlNodes := htmlquery.Find(p.scriptTree, hook.Hook)
			contents := []string{}
			for _, htmlNode := range htmlNodes {
				contents = append(contents, htmlquery.InnerText(htmlNode))
			}
			p.details[hook.Name] = contents
		} else {
			htmlNode := htmlquery.FindOne(p.scriptTree, hook.Hook)
			if hook.Attribute != "" {
				p.details[hook.Name] = htmlquery.SelectAttr(htmlNode, hook.Attribute)
			} else {
				p.details[hook.Name] = htmlquery.InnerText(htmlNode)
			}
		}
	}
}

func (pd *Podcast) ExtractPropertiesFromFeed() error {
	trackNo, err := pd.TrackNo()
	if err != nil {
		return err
	}

	pd.details["trackNo"] = trackNo
	return nil
}

func (p *Podcast) ExtractDirectProperties() {
	p.details["cover"] = p.DirectFields.Cover
	p.details["artist"] = p.DirectFields.Artist
	p.details["album"] = p.DirectFields.Album
	p.details["intro"] = p.DirectFields.IntroURL
}

func (p *Podcast) ExtractProperties() (err error) {
	p.ExtractPropertiesFromScript()
	p.ExtractDirectProperties()
	p.details["pubDate"] = Now()
	p.details["master"] = strings.Replace(p.MasterURLPattern, "<FILE>", p.trackName, 1)
	err = p.ExtractPropertiesFromFeed()

	return err
}
