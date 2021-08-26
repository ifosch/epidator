package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/ifosch/stationery/pkg/gdrive"
	"github.com/ifosch/stationery/pkg/stationery"
)

func GetScript(episodeTag string) (string, error) {
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

func main() {
	log.SetFlags(0)

	content, err := GetScript(ExtractEpisodeTagFromTrack(os.Args[1]))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(content)
}
