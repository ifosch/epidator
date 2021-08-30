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

var GetPubDate = time.Now

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

	err = podcast.DownloadScript()
	if err != nil {
		return nil, err
	}

	err = podcast.DownloadFeed()
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

func (p *Podcast) DownloadScript() error {
	script, err := GetScript(p.EpisodeScriptHook())
	if err != nil {
		return err
	}

	p.scriptTree, err = htmlquery.Parse(strings.NewReader(script))
	if err != nil {
		return err
	}

	return nil
}

func (p *Podcast) DownloadFeed() error {
	feed, err := GetFeed(p.FeedURL)
	if err != nil {
		return err
	}

	p.feedTree, err = htmlquery.Parse(strings.NewReader(feed))
	if err != nil {
		return err
	}

	return nil
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
	p.details["pubDate"] = GetPubDate()
	p.details["master"] = strings.Replace(p.MasterURLPattern, "<FILE>", p.trackName, 1)
	err = p.ExtractPropertiesFromFeed()

	return err
}
