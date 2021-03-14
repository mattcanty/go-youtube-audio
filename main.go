package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/patrickmn/go-cache"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
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
	MimeType      string      `json:"mimeType"`
}

type ViewData struct {
	Error              error
	VideoMetadataItems []VideoMetadata
}

type VideoMetadata struct {
	Title, URL, EscapedURL, Mp4URL, EscapedMp4URL string
}

func main() {
	c := cache.New(24*time.Hour, time.Hour)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.SetHeader("Content-Type", "text/html"))

	r.Get("/{videoID}", func(w http.ResponseWriter, r *http.Request) {
		videoID := chi.URLParam(r, "videoID")

		t, err := template.ParseFiles("index.html")
		handleError(err)

		var b bytes.Buffer
		err = t.Execute(&b, getVideoMetadata(videoID))

		w.Header().Add("Content-Type", "text/html")
		fmt.Fprint(w, b.String())
	})

	r.Get("/channel/{channelID}", func(w http.ResponseWriter, r *http.Request) {
		channelID := chi.URLParam(r, "channelID")
		viewDataCacheKey := fmt.Sprintf("%s-viewdata", channelID)

		if x, found := c.Get(viewDataCacheKey); found {
			viewData := (**x.(**ViewData))

			t, err := template.ParseFiles("channel.html")
			handleError(err)

			var b bytes.Buffer
			err = t.Execute(&b, &viewData)

			w.Header().Add("Content-Type", "text/html")
			fmt.Fprint(w, b.String())
		} else {

			ctx := context.Background()

			youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(os.Getenv("GOOGLE_API_KEY")))
			handleError(err)

			videoList := youtubeService.Videos.List([]string{"snippet"})

			searchList := youtubeService.Search.List([]string{"id"})
			searchList.ChannelId(channelID)
			searchList.Order("date")
			searchList.Type("video")
			searchList.MaxResults(50)
			searchList.PublishedAfter("1970-01-01T00:00:00Z")
			searchResponse, err := searchList.Do()
			handleError(err)

			totalResults := searchResponse.PageInfo.TotalResults
			fmt.Printf("Total Results: %d\n", totalResults)

			var items []*youtube.SearchResult

			items = append(items, searchResponse.Items...)

			fmt.Println(items)

			for {
				fmt.Println("Getting more results")
				lastVideoID := items[len(items)-1].Id.VideoId

				fmt.Printf("Last video ID: %s\n", lastVideoID)
				videoList.Id(lastVideoID)
				videoListResponse, err := videoList.Do()
				handleError(err)

				lastVideo := videoListResponse.Items[0]

				fmt.Printf("Video - %s Title: %s \n", lastVideo.Snippet.PublishedAt, lastVideo.Snippet.Title)
				searchList.PublishedBefore(lastVideo.Snippet.PublishedAt)

				searchResponse2, err := searchList.Do()
				handleError(err)

				items = append(items, searchResponse2.Items[1:]...)

				fmt.Printf("Total Items: %d\n", len(items))

				if len(searchResponse2.Items[1:]) == 0 {
					break
				}
			}

			var viewData = &ViewData{}
			for _, item := range items {
				cacheKey := fmt.Sprintf("%s-videometa", item.Id.VideoId)

				if x, found := c.Get(cacheKey); found {
					videoMetadata := x.(*VideoMetadata)
					viewData.VideoMetadataItems = append(
						viewData.VideoMetadataItems,
						*videoMetadata,
					)
				} else {
					videoMetadata := getVideoMetadata(item.Id.VideoId)
					c.Set(cacheKey, &videoMetadata, cache.DefaultExpiration)
					viewData.VideoMetadataItems = append(
						viewData.VideoMetadataItems,
						videoMetadata,
					)
				}
			}

			c.Set(viewDataCacheKey, &viewData, cache.DefaultExpiration)

			t, err := template.ParseFiles("channel.html")
			handleError(err)

			var b bytes.Buffer
			err = t.Execute(&b, viewData)

			w.Header().Add("Content-Type", "text/html")
			fmt.Fprint(w, b.String())
		}
	})

	http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), r)
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func getVideoMetadata(videoID string) VideoMetadata {
	var client http.Client
	resp, err := client.Get("https://www.youtube.com/get_video_info?video_id=" + videoID)
	handleError(err)
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	handleError(err)
	bodyString := string(bodyBytes)

	kvStrings := strings.Split(bodyString, "&")

	var kvStringMap = make(map[string]string)
	for _, kvString := range kvStrings {
		tmpString := strings.SplitN(kvString, "=", 2)
		kvStringMap[tmpString[0]] = tmpString[1]
	}

	playerResponseJSON, err := url.QueryUnescape(kvStringMap["player_response"])
	handleError(err)

	var playerResponse PlayerResponse
	json.Unmarshal([]byte(playerResponseJSON), &playerResponse)
	lowestContentLength := 0

	var selectedFormat AdaptiveFormat
	var mp4Format AdaptiveFormat
	for _, format := range playerResponse.StreamingData.AdaptiveFormats {
		if strings.Contains(format.MimeType, "audio/mp4") {
			mp4Format = format
		}
		if format.AudioQuality != "AUDIO_QUALITY_LOW" {
			continue
		}
		if lowestContentLength == 0 || selectedFormat.ContentLength > format.ContentLength {
			selectedFormat = format
		}
	}

	meta := &VideoMetadata{
		Title:         playerResponse.VideoDetails.Title,
		URL:           selectedFormat.URL,
		EscapedURL:    url.QueryEscape(selectedFormat.URL),
		Mp4URL:        mp4Format.URL,
		EscapedMp4URL: url.QueryEscape(mp4Format.URL),
	}

	return *meta
}
