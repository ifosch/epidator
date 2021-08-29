package epidactor

import (
	"fmt"
	"testing"
	"time"
)

func MockGetScript(expectedEpisodeHook, expectedOutput string) func(string) (string, error) {
	return func(episodeHook string) (string, error) {
		if episodeHook != expectedEpisodeHook {
			return "", fmt.Errorf("document not found. Query was \"name contains '%s'\"", episodeHook)
		}

		return expectedOutput, nil
	}
}

func MockGetFeed(expectedOutput string) func(string) (string, error) {
	return func(string) (string, error) {
		return expectedOutput, nil
	}
}

func MockNow(fakeTime time.Time) func() time.Time {
	return func() time.Time {
		return fakeTime
	}
}

func TestGetEpisodeDetails(t *testing.T) {
	tc := []struct {
		inputTrackName string
		episodeHook    string
		inputYAMLFile  string
		script         string
		feed           string
		expectedOutput map[string]interface{}
	}{
		{
			"mypodcast-1.master.mp3",
			"Podcast 1",
			"../../propertiesDefinitions.yaml",
			`<html><body>
			<p><span style="font-size:24pt;">Podcast 1 - My Podcast</span></p>
			<p><span><img src="https://my.podcast.com/images/image.png"/></span></p>
			<p style="color:#666666;">In this episode I'll talk about my podcast.</p>
			<ul>
			  <li>One mentioned link - https://my.podcast.com/referrals</li>
			</ul>
			</html></body>`,
			`<rss>
			</rss>`,
			map[string]interface{}{
				"album":  "My Podcast",
				"artist": "Me",
				"cover":  "https://my.podcast.com/media/cover.png",
				"image":  "https://my.podcast.com/images/image.png",
				"intro":  "https://my.podcast.com/media/intro.mp3",
				"links": []string{
					"One mentioned link - https://my.podcast.com/referrals",
				},
				"master":  "https://my.podcast.com/masters/mypodcast-1.master.mp3",
				"pubDate": time.Now(),
				"summary": "In this episode I'll talk about my podcast.",
				"title":   "Podcast 1 - My Podcast",
				"trackNo": 1,
			},
		},
		{
			"mypodcast-colaboracion1.master.mp3",
			"Colaboración 1",
			"../../propertiesDefinitions.yaml",
			`<html><body>
			<p><span style="font-size:24pt;">Collaboration 1 - My friend's episode</span></p>
			<p style="color:#666666;">In this collaboration, I'm hosting my friend's episode.</p>
			<ul>
			<li>One mentioned link - https://my.podcast.com/referrals</li>
			<li>Another mentioned link - https://my.podcast.com/patreon</li>
			</ul>
			</html></body>`,
			`<rss>
			<item></item>
			</rss>`,
			map[string]interface{}{
				"album":  "My Podcast",
				"artist": "Me",
				"cover":  "https://my.podcast.com/media/cover.png",
				"image":  "",
				"intro":  "https://my.podcast.com/media/intro.mp3",
				"links": []string{
					"One mentioned link - https://my.podcast.com/referrals",
					"Another mentioned link - https://my.podcast.com/patreon",
				},
				"master":  "https://my.podcast.com/masters/mypodcast-colaboracion1.master.mp3",
				"pubDate": time.Now(),
				"summary": "In this collaboration, I'm hosting my friend's episode.",
				"title":   "Collaboration 1 - My friend's episode",
				"trackNo": 2,
			},
		},
	}

	for _, tt := range tc {
		GetScript = MockGetScript(tt.episodeHook, tt.script)
		GetFeed = MockGetFeed(tt.feed)
		Now = MockNow(tt.expectedOutput["pubDate"].(time.Time))

		ed, err := GetEpisodeDetails(tt.inputTrackName, tt.inputYAMLFile)
		if err != nil {
			t.Error("Unexpected error:", err)
		} else {
			for k, v := range tt.expectedOutput {
				if k != "links" {
					if ed[k] != v {
						t.Errorf("Unmatching output for %s. Got '%s', expected '%s'", k, ed[k], v)
					}
				} else {
					if len(ed[k].([]string)) != len(v.([]string)) {
						t.Errorf("Unmatching number of elements for %s, %s and %s", k, ed[k], v)
					} else {
						for i, l := range v.([]string) {
							if ed[k].([]string)[i] != l {
								t.Errorf("Unmatching output for %s[%d]. Got '%s', expected '%s'", k, i, ed[k].([]string)[i], l)
							}
						}
					}
				}
			}
		}
	}
}
