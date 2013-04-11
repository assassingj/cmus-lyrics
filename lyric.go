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
)

const API_URL = "http://geci.me/api/lyric"

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
	Status string
	Attrs  map[string]string
}

func GetCurrentSongMetaData() (out string) {
	cmd := exec.Command("cmus-remote", "-Q")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		//not running
		return ""
	}
	return stdout.String()
}

func ParseSongMetaData(out string) SongMeta {
	kvs := strings.Split(out, "\n")
	songMeta := SongMeta{Status: "stop"}
	//status stopped
	if strings.Split(kvs[0], " ")[1] == "stopped" {
		return songMeta
	}
	attr := make(map[string]string)
	for _, kv := range kvs[1:] {
		items := strings.Split(kv, " ")
		fmt.Println(items)
		switch len(items) {
		case 2:
			attr[items[0]] = items[1]
		case 3:
			attr[items[1]] = items[2]
		default:
			log.Println("invalid data while parsing cmus song metadata error,attr:", items)
		}

	}
	songMeta.Attrs = attr
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
	if lyricResult.Code != 0 {
		panic("code error when get lyric:" + url)
	}
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
	// lyric := GetLyric("海阔天空", "信乐团")
	// fmt.Println(lyric)
	out := GetCurrentSongMetaData()
	fmt.Println(ParseSongMetaData(out))
}
