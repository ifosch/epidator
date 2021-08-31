package epidator

func GetEpisodeDetails(trackName, podcastYAML string) (map[string]interface{}, error) {
	podcast, err := NewPodcast(trackName, podcastYAML)
	if err != nil {
		return nil, err
	}

	return podcast.details, nil
}
