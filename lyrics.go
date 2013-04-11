package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const API_URL = "http://geci.me/api/lyric"
const DEFAULT_SLEEP_TIME = 1 * time.Second

const (
	STOPPED = iota
	PLAYING
	ERROR
)

type Lrc struct {
	Url    string `json:"lrc"`
	Song   string `json:"song"`
	Artist string `json:"artist"`
	Aid    int64  `json:"aid"`
	Sid    int64  `json:"sid"`
}

type LyricResult struct {
	Count int64 `json:"count"`
	Code  int64 `json:"code"`
	Lrcs  []Lrc `json:"result"`
}

type SongMeta struct {
	Status int
	Title  string
	Artist string
}

func GetCurrentSongMetaData() SongMeta {
	cmd := exec.Command("cmus-remote", "-Q")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		//not running
		return SongMeta{Status: ERROR}
	}
	songMeta := SongMeta{Status: STOPPED}
	kvs := strings.Split(stdout.String(), "\n")
	//status stopped
	if strings.Split(kvs[0], " ")[1] == "stopped" {
		return songMeta
	}
	songMeta.Status = PLAYING
	attr := make(map[string]string)
	for _, kv := range kvs[1:] {
		if kv = strings.TrimSpace(kv); kv == "" {
			continue
		}
		items := strings.Split(kv, " ")
		if len(items) < 3 || items[0] != "tag" {
			continue
		}
		attr[items[1]] = strings.Join(items[2:], " ")

	}
	songMeta.Title = attr["title"]
	songMeta.Artist = attr["artist"]
	return songMeta
}

func HttpGet(url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	if res.StatusCode != 200 {
		panic("http code error, url=" + url)
	}
	defer res.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return body
}

func GetLyricResult(song, artist string) LyricResult {
	url := fmt.Sprintf("%s/%s", API_URL, song)
	if artist != "" {
		url = fmt.Sprintf("%s/%s", url, artist)
	}

	body := HttpGet(url)
	lyricResult := LyricResult{}
	err := json.Unmarshal(body, &lyricResult)
	if err != nil {
		panic(err)
	}
	// if lyricResult.Code != 0 {
	// 	panic("code error when getting lyric:" + url)
	// }
	if lyricResult.Count <= 0 {
		panic("not found.url:" + url)
	}
	return lyricResult
}

//Only get the first lyric if multiple lyrics found
func GetFirstLyric(song, artist string) string {
	content := ""
	if song == "" {
		return ""
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()

	lyricResult := GetLyricResult(song, artist)
	content = string(HttpGet(lyricResult.Lrcs[0].Url))
	return content
}

func main() {
	currentMeteData := SongMeta{}
	for {
		metaData := GetCurrentSongMetaData()
		if metaData.Status == PLAYING {
			if metaData.Artist != currentMeteData.Artist || metaData.Title != currentMeteData.Title {
				currentMeteData.Artist = metaData.Artist
				currentMeteData.Title = metaData.Title
				fmt.Println("start fetch:", currentMeteData.Title)
				lrc := GetFirstLyric(currentMeteData.Title, currentMeteData.Artist)
				if lrc == "" {
					lrc = GetFirstLyric(currentMeteData.Title, "")
				}
				fmt.Println(lrc)
			}
		}
		time.Sleep(DEFAULT_SLEEP_TIME)
	}
}
