package epidactor

import (
	"io/ioutil"

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
