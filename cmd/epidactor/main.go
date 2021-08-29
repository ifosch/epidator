package main

import (
	"fmt"
	"log"
	"os"

	"github.com/EDyO/epidactor/pkg/epidactor"
	"gopkg.in/yaml.v2"
)

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

	trackName := os.Args[1]
	propertiesDefinitionsYAML := os.Getenv("PROPERTIES_DEFINITIONS_YAML")
	if propertiesDefinitionsYAML == "" {
		propertiesDefinitionsYAML = "propertiesDefinitions.yaml"
	}

	properties, err := epidactor.GetEpisodeDetails(trackName, propertiesDefinitionsYAML)
	if err != nil {
		log.Fatal(err)
	}

	PrintYAML(properties)
}
