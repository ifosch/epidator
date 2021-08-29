package epidactor

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

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
