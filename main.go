package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type PlayerResponse struct {
	StreamingData StreamingData `json:"streamingData"`
}

type StreamingData struct {
	AdaptiveFormats []AdaptiveFormat `json:"adaptiveFormats"`
}

type AdaptiveFormat struct {
	AudioQuality  string      `json:"audioQuality"`
	ContentLength json.Number `json:"contentLength"`
	URL           string      `json:"url"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.SetHeader("Content-Type", "text/html"))

	r.Get("/{videoID}", func(w http.ResponseWriter, r *http.Request) {
		videoID := chi.URLParam(r, "videoID")

		var client http.Client
		resp, err := client.Get("https://www.youtube.com/get_video_info?video_id=" + videoID)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)

		kvStrings := strings.Split(bodyString, "&")

		var kvStringMap = make(map[string]string)
		for _, kvString := range kvStrings {
			tmpString := strings.SplitN(kvString, "=", 2)
			kvStringMap[tmpString[0]] = tmpString[1]
		}

		playerResponseJSON, err := url.QueryUnescape(kvStringMap["player_response"])
		if err != nil {
			log.Fatal(err)
		}

		var playerResponse PlayerResponse
		json.Unmarshal([]byte(playerResponseJSON), &playerResponse)
		lowestContentLength := 0

		var selectedFormat AdaptiveFormat
		for _, format := range playerResponse.StreamingData.AdaptiveFormats {
			if format.AudioQuality != "AUDIO_QUALITY_LOW" {
				continue
			}
			if lowestContentLength == 0 || selectedFormat.ContentLength > format.ContentLength {
				selectedFormat = format
			}
		}

		fmt.Println(selectedFormat.URL)

		t, err := template.New("foo").Parse(`{{define "T"}}<audio src="{{.}}" controls></audio>{{end}}`)

		err = t.ExecuteTemplate(w, "T", selectedFormat.URL)
	})

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
}
