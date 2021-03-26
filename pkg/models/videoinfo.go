package models

import (
	"encoding/json"
	"net/url"
	"strings"
)

type VideoInfo struct {
	PlayerResponse PlayerResponse `mapstructure:"player_response"`
}

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

func Decode(videoInfoData string, videoInfo *VideoInfo) error {
	for _, item := range strings.Split(videoInfoData, "&") {
		kv := strings.SplitN(item, "=", 2)
		switch kv[0] {
		case "player_response":
			jsonData, err := url.QueryUnescape(kv[1])
			if err != nil {
				return err
			}

			err = json.Unmarshal([]byte(jsonData), &videoInfo.PlayerResponse)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
