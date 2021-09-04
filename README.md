# Episode Data Extractor

This program extracts data for episodes.

## Requirements

You'll need to get credentials for a [service account](https://support.google.com/a/answer/7378726?hl=en) with access to the documents you want to use on Google Drive. The path to the credentials file is to be supplied as the `DRIVE_CREDENTIALS_FILE` environment variable.

## Install
### Downloadable binaries

To get compiled binaries for Windows, MacOS, and Linux, on amd64 architecture, you can go to any of our [releases](https://github.com/EDyO/epidator/releases).

### From source code

To setup these tools from source code, you'll need to get [Go language development environment](https://golang.org/doc/install) up and running on a Linux computer. Then, you'll need to get the code, either by cloning or downloading a compressed copy of the code and uncompress it. Once done, you can use the `scripts/build.sh` to build the binary to whichever operating system and architecture you want:

```bash
scripts/build.sh windows amd64
```

Once this is completed, you'll find the binary in the build directory.
Finally, you can move it to whatever in your execution path.

## Setup

The main configuration file must be defined in the file pointed by the `PODCAST_YAML` environment variable.
By default, it looks for `podcast.yaml`.

This file has the following possible lines:
- feedURL: This must be the URL for the podcast's feed
- masterURLPattern: This must be the URL where the master can be downloaded from.
  Use `<FILE>` as replacement mark for the audio track name
- directFields: These fields are direct properties which won't be modified by the program:
  - introURL: This must be the URL for the podcast's feed intro music
  - cover: This must be the URL for the podcast's cover image
  - artist: This must be the name for the podcast's author ID3 tag
  - album: This must be the name for the podcast's album ID3 tag
- scriptFieldHooks: This is a list of extracting descriptions of fields from the HTML script.
  The list can be empty, but if any is present, the following fields need to be present:
  - name: Corresponds to the field name to fill in the resulting YAML
  - hook: This must be an XPath search string to identify the value from the script for this field
  - list: false
- episodeScriptHooks: This is a map (`key: value`) of different types of episodes (podcast, pills, etc.)
- episodeBucket: This is the bucket name where all the published tracks are stored
  The program uses the key as a pattern match on the track name, along with the number, to identify the script.
  The `default` key must be present.

## Usage

The command `epidator` must receive an argument which is the file name of the audio track available in `masterURLPattern` property from the `PODCAST_YAML` file.
Using this file name and the properties defined in the `PODCAST_YAML` file, it will extract the properties of this episode in order to be edited and published.
The output is printed in YAML format:

```bash
$ epidator episode-1.master.mp3
album: My Podcast
artist: Me
cover: https://my.podcast.com/media/cover.png
intro: https://my.podcast.com/media/intro.mp3
master: https://my.podcast.com/masters/episode-1.master.mp3
pubDate: 2021-08-31T20:52:50.658906773+02:00
trackNo: 1
bucket: mypodcast-episodes
```
