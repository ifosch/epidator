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
	podcastYAML := os.Getenv("PODCAST_YAML")
	if podcastYAML == "" {
		podcastYAML = "podcast.yaml"
	}

	details, err := epidactor.GetEpisodeDetails(trackName, podcastYAML)
	if err != nil {
		log.Fatal(err)
	}

	PrintYAML(details)
}
