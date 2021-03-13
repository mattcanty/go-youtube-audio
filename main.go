package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/cavaliercoder/grab"

	ffmpeg "github.com/floostack/transcoder/ffmpeg"
)

// PlayerResponse contains all the details we need to get audio
type PlayerResponse struct {
	StreamingData StreamingData `json:"streamingData"`
	VideoDetails  VideoDetails  `json:"videoDetails"`
}

type VideoDetails struct {
	Title string `json:"title"`
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

		t, err := template.ParseFiles("index.html")
		if err != nil {
			log.Fatal(err)
		}

		var b bytes.Buffer
		err = t.Execute(&b, struct {
			Title, URL, EscapedURL string
		}{
			Title:      playerResponse.VideoDetails.Title,
			URL:        selectedFormat.URL,
			EscapedURL: url.QueryEscape(selectedFormat.URL),
		})

		w.Header().Add("Content-Type", "text/html")
		fmt.Fprint(w, b.String())
	})

	r.Get("/mp3/{videoURL}", func(w http.ResponseWriter, r *http.Request) {
		url, err := url.QueryUnescape(chi.URLParam(r, "videoURL"))
		if err != nil {
			log.Fatal(err)
		}
		urlHash := fnv.New32a()
		urlHash.Write([]byte(url))

		inputFilePath := fmt.Sprintf("/tmp/%d.webm", urlHash.Sum32())
		outputFilePath := fmt.Sprintf("/tmp/%d.mp3", urlHash.Sum32())

		fmt.Printf("Downloading %s...\n", url)
		_, err = grab.Get(inputFilePath, url)
		if err != nil {
			log.Fatal(err)
		}

		format := "mp3"
		overwrite := true
		opts := ffmpeg.Options{
			OutputFormat: &format,
			Overwrite:    &overwrite,
		}

		ffmpegConf := &ffmpeg.Config{
			FfmpegBinPath:   "ffmpeg",
			FfprobeBinPath:  "ffprobe",
			ProgressEnabled: true,
		}

		progress, err := ffmpeg.
			New(ffmpegConf).
			InputPipe().
			Input(inputFilePath).
			Output(outputFilePath).
			WithOptions(opts).
			Start(opts)

		if err != nil {
			log.Fatal(err)
		}

		for msg := range progress {
			log.Printf("%+v", msg)
		}

		file, err := os.Open(outputFilePath)
		if err != nil {
			log.Fatal(err)
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Disposition", "filename=audio.mp3")
		w.Write(data)
	})

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
}
