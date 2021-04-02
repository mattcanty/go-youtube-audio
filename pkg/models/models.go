package models

import "encoding/json"

type VideoInfo struct {
	PlayerResponse PlayerResponse
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
